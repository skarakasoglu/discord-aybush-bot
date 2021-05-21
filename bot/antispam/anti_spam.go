package antispam

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"time"
)

const (
	MAX_CACHE_DURATION = 600
)

type guildMessages map[string]*cachedMemberMessages

type cachedMemberMessages struct {
	messages []*discordgo.Message
}

type cachedChanMessage struct{
	index int
	*cachedMemberMessages
}

type AntiSpam struct{
	running bool

	options
	cachedMessages map[string]guildMessages

	onMessageChan chan *discordgo.Message
	cachedMessageChan chan *cachedChanMessage
}

func NewAntiSpam(maxInterval int, maxDuplicatesInterval int,
	exemptRoles []string, ignoredUsers []string, ignoredChannels []string) AntiSpam {
	return AntiSpam{
		options: options{
			maxInterval:      maxInterval,
			maxDuplicatesInterval: maxDuplicatesInterval,
			exemptRoles:      exemptRoles,
			ignoredUsers:     ignoredUsers,
			ignoredChannels: ignoredChannels,
		},
		cachedMessages: make(map[string]guildMessages),
		onMessageChan: make(chan *discordgo.Message, 1000),
		cachedMessageChan: make(chan *cachedChanMessage, 1000),
	}
}

func (antiSpam *AntiSpam) AddProtectionConfig(config ProtectionConfig) {
	antiSpam.protectionConfigurations = append(antiSpam.protectionConfigurations, config)
}

func (antiSpam *AntiSpam) Start() {
	antiSpam.running = true

	go antiSpam.workAsync()
}

func (antiSpam *AntiSpam) Stop() {
	antiSpam.running = false
}

func (antiSpam *AntiSpam) OnMessage(message *discordgo.Message) {
	antiSpam.onMessageChan <- message
}

func (antiSpam *AntiSpam) workAsync() {
	for antiSpam.running {
		select {
		case message := <- antiSpam.onMessageChan:
			antiSpam.messageReceived(message)
		case cachedMessage := <- antiSpam.cachedMessageChan:
			cachedMessage.messages = append(cachedMessage.messages[:cachedMessage.index], cachedMessage.messages[cachedMessage.index + 1:]...)
		}
	}
}

func (antiSpam *AntiSpam) messageReceived(message *discordgo.Message) {
	if antiSpam.shouldIgnore(message) {
		return
	}

	_, ok := antiSpam.cachedMessages[message.GuildID]
	if !ok {
		antiSpam.cachedMessages[message.GuildID] = make(guildMessages)
	}

	_, ok = antiSpam.cachedMessages[message.GuildID][message.Author.ID]
	if !ok {
		antiSpam.cachedMessages[message.GuildID][message.Author.ID] = &cachedMemberMessages{}
	}

	memberMessages, _ := antiSpam.cachedMessages[message.GuildID][message.Author.ID]

	memberMessages.messages = append(memberMessages.messages, message)
	messageIndex := len(memberMessages.messages) - 1

	// Delete cached messages after a certain amount of time.
	go func(index int) {
		time.Sleep(time.Duration(MAX_CACHE_DURATION) * time.Second)
		antiSpam.cachedMessageChan <- &cachedChanMessage{
			index:    index,
			cachedMemberMessages: memberMessages,
		}
	}(messageIndex)

	for _, protection := range antiSpam.protectionConfigurations {
		var spamMatches []*discordgo.Message
		var duplicateMatches []*discordgo.Message


		lastMessageTime, err := message.Timestamp.Parse()
		if err != nil {
			log.Printf("[AybushBot::AntiSpam] Error on parsing timestamp data for last message: %v", err)
			lastMessageTime = time.Now()
		}

		for _, memberMessage := range memberMessages.messages {
			sentTime, err := memberMessage.Timestamp.Parse()
			if err != nil {
				log.Printf("[AybushBot::AntiSpam] Error on parsing timestamp data for a member message: %v", err)
				continue
			}

			lastMessageMilli := lastMessageTime.UnixNano() / int64(time.Millisecond)
			sentMilli := sentTime.UnixNano() / int64(time.Millisecond)

			if lastMessageMilli - sentMilli <  int64(antiSpam.maxInterval) {
				spamMatches = append(spamMatches, memberMessage)
			}

			if (time.Now().UnixNano() / int64(time.Millisecond)) - (sentTime.UnixNano() / int64(time.Millisecond)) <  int64(antiSpam.maxDuplicatesInterval) &&
				memberMessage.Content == message.Content {
				duplicateMatches = append(duplicateMatches, memberMessage)
			}

		}

		if len(spamMatches) >= protection.Threshold || len(duplicateMatches) >= protection.MaxDuplicates {
			log.Printf("[AybushBot::AntiSpam] Spam detected in guild %v. Member: %v, spam matches: %v, duplicateMatches: %v",
				message.GuildID, message.Author.ID, len(spamMatches), len(duplicateMatches))
			i := 0
			for _, memberMessage := range memberMessages.messages {
				for _, spamMessage := range spamMatches {
					if spamMessage.ID == memberMessage.ID {
						memberMessages.messages[i] = memberMessages.messages[0]
						memberMessages.messages = memberMessages.messages[1:]
						i--
					}
				}
				i++
			}

			protection.Callback(message.GuildID, message.Author.ID, spamMatches)
		}
	}
}

func (antiSpam *AntiSpam) shouldIgnore(message *discordgo.Message) bool {
	isRoleIgnored := func() bool {
		if antiSpam.exemptRoles == nil {
			return false
		}

		if message.Member != nil {
			for _, role := range antiSpam.exemptRoles {
				for _, memberRole := range message.Member.Roles {
					if memberRole == role {
						return true
					}
				}
			}
		}

		return false
	}()

	isUserIgnored := func() bool {
		if antiSpam.ignoredUsers == nil {
			return false
		}

		for _, userId := range antiSpam.ignoredUsers {
			if userId == message.Author.ID {
				return true
			}
		}

		return false
	}()

	isChannelIgnored := func() bool {
		if antiSpam.ignoredChannels == nil {
			return false
		}

		for _, channelId := range antiSpam.ignoredChannels {
			if channelId == message.ChannelID {
				return true
			}
		}

		return false
	}()

	return isRoleIgnored || isUserIgnored || isChannelIgnored
}