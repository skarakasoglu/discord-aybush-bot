package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/skarakasoglu/discord-aybush-bot/configuration"
	"log"
	"strconv"
	"strings"
	"time"
)

type muteCommand struct{
	session *discordgo.Session
}

func NewMuteCommand(session *discordgo.Session) Command{
	return &muteCommand{
		session: session,
	}
}

func (cmd *muteCommand) Name() string{
	return "mute"
}

func (cmd *muteCommand) Execute(message *discordgo.Message) (string, error) {
	arguments := strings.Split(message.Content, " ")[1:]

	isAuthorized := func() bool {
		for _, role := range configuration.Manager.Roles.ModerationRoles {
			for _, memberRole := range message.Member.Roles {
				if memberRole == role {
					return true
				}
			}
		}
		return false
	}()

	if !isAuthorized {
		botLogMsg := fmt.Sprintf("> <@%v> yetkisi olmayan \"**%v**\" komutunu kullanmaya çalıştı.", message.Author.ID, cmd.Name())
		_, err := cmd.session.ChannelMessageSend(configuration.Manager.Channels.BotLog, botLogMsg)
		if err != nil {
			log.Printf("Error on writing log to bot log channel: %v", err)
		}

		err = cmd.session.ChannelMessageDelete(message.ChannelID, message.ID)
		if err != nil {
			log.Printf("Error on deleting command message: %v", err)
		}
		return fmt.Sprintf("> <@%v>, bu komutu kullanma **yetkiniz** bulunmamaktadır.", message.Author.ID), nil
	}

	if len(arguments) != 2 || len(message.Mentions) != 1 {
		return fmt.Sprintf("<@%v>, %v", message.Author.ID, cmd.Usage()), nil
	}

	durationArg := arguments[1]
	durationType := durationArg[len(durationArg) - 1:]

	var durationMultiplier time.Duration
	if durationType == "h" {
		durationMultiplier = time.Hour
	} else if durationType == "m" {
		durationMultiplier = time.Minute
	} else if durationType == "s" {
		durationMultiplier = time.Second
	} else {
		log.Printf("Invalid duration type: %v", durationType)
		return fmt.Sprintf("<@%v>, %v", message.Author.ID, cmd.Usage()), nil
	}

	duration, err := strconv.Atoi(durationArg[:len(durationArg) - 1])
	if err != nil {
		log.Printf("Error on converting duration to integer: %v", err)
		return fmt.Sprintf("<@%v>, %v", message.Author.ID, cmd.Usage()), nil
	}

	mutedMember := message.Mentions[0]
	member,err := cmd.session.GuildMember(message.GuildID, mutedMember.ID)
	if err != nil {
		log.Printf("Error on optaining member: %v", err)
		return "", err
	}
	isModerator := func() bool {
		for _, role := range configuration.Manager.Roles.ModerationRoles {
			for _, memberRole := range member.Roles {
				if memberRole == role {
					return true
				}
			}
		}
		return false
	}()
	if isModerator {
		log.Printf("Moderators can not be muted, user: %v",member.User.Username)
		botLogMsg := fmt.Sprintf("> <@%v>, <@%v> kullanıcısına mute rolü veremedi. sebep: **#ModeratörlerSUSTURULAMAZ**", message.Author.ID, mutedMember.ID)
		_, err = cmd.session.ChannelMessageSend(configuration.Manager.Channels.BotLog, botLogMsg)
		if err != nil {
			log.Printf("Error on sending log message to bot log channel: %v", err)
		}
	}else {
		err = cmd.session.GuildMemberRoleAdd(message.GuildID, mutedMember.ID, configuration.Manager.Roles.MuteRole)
		if err != nil {
			log.Printf("Error on adding mute role to a member: %v", err)
			return "", err
		}
		log.Printf("%v#%v muted %v#%v for %v.", message.Author.Username, message.Author.Discriminator,
			mutedMember.Username, mutedMember.Discriminator, durationArg)

		go func() {
			time.Sleep(time.Duration(duration) * durationMultiplier)

			err = cmd.session.GuildMemberRoleRemove(message.GuildID, mutedMember.ID, configuration.Manager.Roles.MuteRole)
			if err != nil {
				log.Printf("Error on removing muted role: %v", err)
			}

			botLogMsg := fmt.Sprintf("> <@%v> kullanıcısının **%v** sürelik mute rolü kaldırıldı.", mutedMember.ID, durationArg)
			_, err = cmd.session.ChannelMessageSend(configuration.Manager.Channels.BotLog, botLogMsg)

			if err != nil {
				log.Printf("Error on writing log to bot log channel: %v", err)
			}

			dmChannel, err := cmd.session.UserChannelCreate(mutedMember.ID)
			if err != nil {
				log.Printf("Error on creating DM channel with muted user: %v", err)
			}
			if dmChannel == nil {
				log.Printf("DM Channel not created")
			}
			_, err = cmd.session.ChannelMessageSend(dmChannel.ID, fmt.Sprintf("> Sunucuda **%v** sürelik susturulman kaldırıldı.", durationArg))
			if err != nil {
				log.Printf("Error sending message to member via DM Channel: %v", err)
			}

		}()

		botLogMsg := fmt.Sprintf("> <@%v>, <@%v> kullanıcısına **%v** mute rolü verdi.", message.Author.ID, mutedMember.ID, durationArg)
		_, err = cmd.session.ChannelMessageSend(configuration.Manager.Channels.BotLog, botLogMsg)
		if err != nil {
			log.Printf("Error on sending log message to bot log channel: %v", err)
		}

		dmChannel, err := cmd.session.UserChannelCreate(mutedMember.ID)
		if err != nil {
			log.Printf("Error on creating DM channel with muted user: %v", err)
			return "", nil
		}

		_, err = cmd.session.ChannelMessageSend(dmChannel.ID, fmt.Sprintf("> Sunucuda **%v** süreliğine susturuldun.", durationArg))
		if err != nil {
			log.Printf("Error sending message to member via DM Channel: %v", err)
		}
	}
	err = cmd.session.ChannelMessageDelete(message.ChannelID, message.ID)
	if err != nil {
		log.Printf("Error on deleting command message: %v", err)
	}

	return "", nil
}

func (cmd *muteCommand) Usage() string{
	return "**bu komut,** \n\"\\>!mute `<kullanıcı-adı>` `<süre(h|m|s)>`\"\n şeklinde kullanılır. *(Moderasyon yetkisi gerektirir.)*"
}
