package antispam

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"sync"
	"time"
)

const (
	MAX_CACHE_DURATION = 600
)

type guildMessages map[string]cachedMemberMessages

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
	//Protection specific variables
	muteThreshold int
	muteCallBack func(string, string, []*discordgo.Message)

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

func (antiSpam AntiSpam) OnMessage(message *discordgo.Message) {
	isRoleIgnored := func() bool {
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
		for _, userId := range antiSpam.ignoredUsers {
			if userId == message.Author.ID {
				return true
			}
		}

		return false
	}()

	if isRoleIgnored ||isUserIgnored {
		return
	}

	_, ok := antiSpam.cachedMessages[message.GuildID]
	if !ok {
		antiSpam.cachedMessages[message.GuildID] = make(guildMessages)
	}

	_, ok = antiSpam.cachedMessages[message.GuildID][message.Author.ID]
	if !ok {
		antiSpam.cachedMessages[message.GuildID][message.Author.ID] = cachedMemberMessages{}
	}

	antiSpam.guildMessagesMtx.Lock()
	memberMessages, _ := antiSpam.cachedMessages[message.GuildID][message.Author.ID]

	memberMessages.messageMtx.Lock()
	memberMessages.messages = append(memberMessages.messages, message)
	memberMessages.messageMtx.Unlock()

	antiSpam.cachedMessages[message.GuildID][message.Author.ID] = memberMessages
	antiSpam.guildMessagesMtx.Unlock()

	// Delete cached messages after a certain amount of time.
	go func() {
		time.Sleep(time.Duration(MAX_CACHE_DURATION) * time.Second)

		for i, val := range memberMessages.messages {
			if val == message {
				memberMessages.messageMtx.Lock()
				memberMessages.messages = append(memberMessages.messages[:i], memberMessages.messages[i + 1:]...)
				memberMessages.messageMtx.Unlock()

				antiSpam.guildMessagesMtx.Lock()
				antiSpam.cachedMessages[message.GuildID][message.Author.ID] = memberMessages
				antiSpam.guildMessagesMtx.Unlock()

				break
			}
		}
	}()

	for _, protection := range antiSpam.protectionConfigurations {
		var spamMatches []*discordgo.Message
		var duplicateMatches []*discordgo.Message

		fmt.Println(protection)
		memberMessages.messageMtx.Lock()
		for _, memberMessage := range memberMessages.messages {
			sentTime, err := memberMessage.Timestamp.Parse()
			if err != nil {
				log.Printf("Error on parsing timestamp data for a member message: %v", err)
			}

			if (time.Now().UnixNano() / int64(time.Millisecond)) - (sentTime.UnixNano() / int64(time.Millisecond)) <  int64(antiSpam.maxInterval) {
				spamMatches = append(spamMatches, memberMessage)
			}

			if (time.Now().UnixNano() / int64(time.Millisecond)) - (sentTime.UnixNano() / int64(time.Millisecond)) <  int64(antiSpam.maxDuplicatesInterval) {
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

