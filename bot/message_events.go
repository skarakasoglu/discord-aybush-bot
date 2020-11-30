package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/skarakasoglu/discord-aybush-bot/configuration"
	"log"
	"mvdan.cc/xurls"
	"time"
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

func (a *Aybus) onSpamCheck(session *discordgo.Session, messageCreate *discordgo.MessageCreate) {
	a.antiSpam.OnMessage(messageCreate.Message)
}

func (a *Aybus) muteUserOnSpam(guildId string, memberId string, spamMessages []*discordgo.Message) {
	log.Printf("User %v muted in guild %v.", memberId, guildId)

	err := a.discordConnection.GuildMemberRoleAdd(guildId, memberId, configuration.Manager.Roles.MuteRole)
	if err != nil {
		log.Printf("Error on adding muted role to member: %v", err)
	}

	lastChannelId := ""
	for _, msg := range spamMessages {
		lastChannelId = msg.ChannelID

		err = a.discordConnection.ChannelMessageDelete(msg.ChannelID, msg.ID)
		if err != nil {
			log.Printf("Error on deleting a spam message: %v", err)
		}
	}

	// After time passes, mute role will be removed from the guild member.
	go func() {
		time.Sleep(time.Duration(configuration.Manager.AntiSpam.Mute.Duration) * time.Millisecond)
		err = a.discordConnection.GuildMemberRoleRemove(guildId, memberId, configuration.Manager.Roles.MuteRole)
		if err != nil {
			log.Printf("Error on removing muted role from member: %v", err)
		}
	}()

	muteDurationInMutes := configuration.Manager.AntiSpam.Mute.Duration / 60000
	notificationMessageToGuildChannel := fmt.Sprintf(configuration.Manager.AntiSpam.Mute.ChannelMessage, memberId, muteDurationInMutes)
	_, err = a.discordConnection.ChannelMessageSend(lastChannelId, notificationMessageToGuildChannel)
	if err != nil {
		log.Printf("Error on sending mute notification message to guild channel: %v", err)
	}

	dmChannel, err := a.discordConnection.UserChannelCreate(memberId)
	if err != nil {
		log.Printf("Error on creating DM channel: %v", err)
		return
	}

	notificationMessageToDMChannel := fmt.Sprintf(configuration.Manager.AntiSpam.Mute.Message, muteDurationInMutes)
	_, err = a.discordConnection.ChannelMessageSend(dmChannel.ID, notificationMessageToDMChannel)
	if err != nil{
		log.Printf("Error on sending mute notification message to DM channel: %v", err)
	}
}