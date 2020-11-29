package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/skarakasoglu/discord-aybush-bot/configuration"
	"log"
	"mvdan.cc/xurls"
)

func (a *Aybus) onURLSend(session *discordgo.Session, messageCreate *discordgo.MessageCreate) {
	isChannelRestricted := func() bool {
		for _, val := range configuration.Manager.UrlRestriction.RestrictedChannels {
			if messageCreate.ChannelID == val {
				return true
			}
		}
		return false
	}()

	if !isChannelRestricted {
		return
	}

	isRoleRestricted := func() bool {
		for _, role := range configuration.Manager.Roles.ModerationRoles {
			for _, memberRole := range messageCreate.Member.Roles {
				if memberRole == role {
					return false
				}
			}
		}

		return true
	}()

	if !isRoleRestricted {
		return
	}

	rxStrict := xurls.Strict()
	urls := rxStrict.FindAllString(messageCreate.Message.Content, -1)

	if len(urls) < 1 {
		return
	}

	msg := fmt.Sprintf("<@%v>, %v", messageCreate.Message.Author.ID, configuration.Manager.UrlRestriction.WarningMessage)
	_, err := session.ChannelMessageSend(messageCreate.ChannelID, msg)
	if err != nil {
		log.Printf("Error on sending warning message to channel: %v", err)
	}

	err = session.ChannelMessageDelete(messageCreate.ChannelID, messageCreate.Message.ID)
	if err != nil {
		log.Printf("Error on deleting a message which contains a URL: %v", err)
	}
}
