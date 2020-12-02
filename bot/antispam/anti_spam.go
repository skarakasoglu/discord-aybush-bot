package antispam

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"sync"
	"time"
)

const (
	MAX_CACHE_DURATION = 600
)

type guildMessages map[string]*cachedMemberMessages

type cachedMemberMessages struct {
	messageMtx sync.Mutex
	messages []*discordgo.Message
}

type ProtectionConfig struct{
	Threshold int
	MaxDuplicates int
	Callback func(string, string, []*discordgo.Message)
}

type options struct{
	protectionConfigurations []ProtectionConfig

	//General variables
	maxInterval int
	maxDuplicatesInterval int
	exemptRoles []string
	ignoredUsers []string
	ignoredChannels []string
}

type AntiSpam struct{
	options

	guildMessagesMtx sync.Mutex
	cachedMessages map[string]guildMessages
}

func NewAntiSpam(maxInterval int, maxDuplicatesInterval int,
	exemptRoles []string, ignoredUsers []string) AntiSpam {
	return AntiSpam{
		options: options{
			maxInterval:      maxInterval,
			maxDuplicatesInterval: maxDuplicatesInterval,
			exemptRoles:      exemptRoles,
			ignoredUsers:     ignoredUsers,
		},
		cachedMessages: make(map[string]guildMessages),
	}
}

func (antiSpam *AntiSpam) AddProtectionConfig(config ProtectionConfig) {
	antiSpam.protectionConfigurations = append(antiSpam.protectionConfigurations, config)
}

func (antiSpam *AntiSpam) OnMessage(message *discordgo.Message) {
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

	memberMessages.messageMtx.Lock()
	memberMessages.messages = append(memberMessages.messages, message)
	memberMessages.messageMtx.Unlock()

	// Delete cached messages after a certain amount of time.
	go func() {
		time.Sleep(time.Duration(MAX_CACHE_DURATION) * time.Second)

		for i, val := range memberMessages.messages {
			if val == message {
				memberMessages.messageMtx.Lock()
				memberMessages.messages = append(memberMessages.messages[:i], memberMessages.messages[i + 1:]...)
				memberMessages.messageMtx.Unlock()
				break
			}
		}
	}()

	for _, protection := range antiSpam.protectionConfigurations {
		var spamMatches []*discordgo.Message
		var duplicateMatches []*discordgo.Message

		memberMessages.messageMtx.Lock()
		for _, memberMessage := range memberMessages.messages {
			sentTime, err := memberMessage.Timestamp.Parse()
			if err != nil {
				log.Printf("Error on parsing timestamp data for a member message: %v", err)
				continue
			}

			if (time.Now().UnixNano() / int64(time.Millisecond)) - (sentTime.UnixNano() / int64(time.Millisecond)) <  int64(antiSpam.maxInterval) {
				spamMatches = append(spamMatches, memberMessage)
			}

			if (time.Now().UnixNano() / int64(time.Millisecond)) - (sentTime.UnixNano() / int64(time.Millisecond)) <  int64(antiSpam.maxDuplicatesInterval) &&
				memberMessage.Content == message.Content {
				duplicateMatches = append(duplicateMatches, memberMessage)
			}

		}
		memberMessages.messageMtx.Unlock()

		if len(spamMatches) >= protection.Threshold || len(duplicateMatches) >= protection.MaxDuplicates {
			log.Printf("Spam detected in guild %v. Member: %v, spam matches: %v", message.GuildID, message.Author.ID, len(spamMatches))
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