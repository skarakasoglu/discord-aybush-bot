package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	embed "github.com/clinet/discordgo-embed"
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
			log.Printf("[AybushBot::MuteUser] Error on writing log to bot log channel: %v", err)
		}

		err = cmd.session.ChannelMessageDelete(message.ChannelID, message.ID)
		if err != nil {
			log.Printf("[AybushBot::MuteUser] Error on deleting command message: %v", err)
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
		log.Printf("[AybushBot::MuteUser] Invalid duration type: %v", durationType)
		return fmt.Sprintf("<@%v>, %v", message.Author.ID, cmd.Usage()), nil
	}

	duration, err := strconv.Atoi(durationArg[:len(durationArg) - 1])
	if err != nil {
		log.Printf("[AybushBot::MuteUser] Error on converting duration to integer: %v", err)
		return fmt.Sprintf("<@%v>, %v", message.Author.ID, cmd.Usage()), nil
	}

	mutedMember := message.Mentions[0]
	member,err := cmd.session.GuildMember(message.GuildID, mutedMember.ID)
	if err != nil {
		log.Printf("[AybushBot::MuteUser] Error on optaining member: %v", err)
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
		log.Printf("[AybushBot::MuteUser] Moderators can not be muted, user: %v",member.User.Username)
		botLogMsg := fmt.Sprintf("> <@%v>, <@%v> kullanıcısına mute rolü veremedi. sebep: **#ModeratörlerSUSTURULAMAZ**", message.Author.ID, mutedMember.ID)
		_, err = cmd.session.ChannelMessageSend(configuration.Manager.Channels.BotLog, botLogMsg)
		if err != nil {
			log.Printf("[AybushBot::MuteUser] Error on sending log message to bot log channel: %v", err)
		}
	}else {
		err = cmd.session.GuildMemberRoleAdd(message.GuildID, mutedMember.ID, configuration.Manager.Roles.MuteRole)
		if err != nil {
			log.Printf("[AybushBot::MuteUser] Error on adding mute role to a member: %v", err)
			return "", err
		}
		log.Printf("%v#%v muted %v#%v for %v.", message.Author.Username, message.Author.Discriminator,
			mutedMember.Username, mutedMember.Discriminator, durationArg)

		go func() {
			time.Sleep(time.Duration(duration) * durationMultiplier)

			err = cmd.session.GuildMemberRoleRemove(message.GuildID, mutedMember.ID, configuration.Manager.Roles.MuteRole)
			if err != nil {
				log.Printf("[AybushBot::MuteUser] Error on removing muted role: %v", err)
			}

			botLogMsg := fmt.Sprintf("> <@%v> kullanıcısının **%v** sürelik susturması kaldırıldı.", mutedMember.ID, durationArg)
			_, err = cmd.session.ChannelMessageSend(configuration.Manager.Channels.BotLog, botLogMsg)

			if err != nil {
				log.Printf("[AybushBot::MuteUser] Error on writing log to bot log channel: %v", err)
			}

			dmChannel, err := cmd.session.UserChannelCreate(mutedMember.ID)
			if err != nil {
				log.Printf("[AybushBot::MuteUser] Error on creating DM channel with muted user: %v", err)
				return
			}

			_, err = cmd.session.ChannelMessageSend(dmChannel.ID, fmt.Sprintf("> Sunucuda **%v** sürelik susturulman kaldırıldı.", durationArg))
			if err != nil {
				log.Printf("[AybushBot::MuteUser] Error sending message to member via DM Channel: %v", err)
			}

		}()

		botLogEmbedMsg := embed.NewGenericEmbed("Moderasyon İşlemi", "")
		botLogEmbedMsg.Color = 0xF97100
		botLogEmbedMsg.Fields = []*discordgo.MessageEmbedField{
			{
				Name:   "İşlem",
				Value:  "Susturma",
				Inline: true,
			},
			{
				Name:   "Uygulanan Kişi",
				Value:  fmt.Sprintf("<@%v>", mutedMember.ID),
				Inline: true,
			},
			{
				Name:   "Kanal",
				Value:  fmt.Sprintf("<#%v>", message.ChannelID),
				Inline: true,
			},
			{
				Name:   "Uygulayan",
				Value:  fmt.Sprintf("<@%v>", message.Author.ID),
				Inline: false,
						},
			{
				Name:   "Süre",
				Value:  fmt.Sprintf("%v", durationArg),
				Inline: false,
						},
		}
		botLogEmbedMsg.Footer = &discordgo.MessageEmbedFooter{
			Text:         time.Now().Format(time.Stamp),
		}

		_, err = cmd.session.ChannelMessageSendEmbed(configuration.Manager.Channels.BotLog, botLogEmbedMsg)
		if err != nil {
			log.Printf("[AybushBot::MuteUser] Error on sending log message to bot log channel: %v", err)
		}

		dmChannel, err := cmd.session.UserChannelCreate(mutedMember.ID)
		if err != nil {
			log.Printf("[AybushBot::MuteUser] Error on creating DM channel with muted user: %v", err)
			return "", nil
		}

		_, err = cmd.session.ChannelMessageSend(dmChannel.ID, fmt.Sprintf("> Sunucuda **%v** süreliğine susturuldun.", durationArg))
		if err != nil {
			log.Printf("[AybushBot::MuteUser] Error sending message to member via DM Channel: %v", err)
		}
	}
	err = cmd.session.ChannelMessageDelete(message.ChannelID, message.ID)
	if err != nil {
		log.Printf("[AybushBot::MuteUser] Error on deleting command message: %v", err)
	}

	return cmd.ResponseMessage(), nil
}

func (cmd* muteCommand) ResponseMessage() string{
	return ""
}

func (cmd *muteCommand) Usage() string{
	return "**bu komut,** \n> \"!mute `<kullanıcı-adı>` `<süre(h|m|s)>`\"\n şeklinde kullanılır. *(Moderasyon yetkisi gerektirir.)*"
}
