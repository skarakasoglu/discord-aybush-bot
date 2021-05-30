package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	embed "github.com/clinet/discordgo-embed"
	"github.com/skarakasoglu/discord-aybush-bot/configuration"
	"github.com/skarakasoglu/discord-aybush-bot/data/models"
	"log"
	"mvdan.cc/xurls/v2"
	"strings"
	"time"
)

func (a *Aybush) saveToDatabase(session *discordgo.Session, create *discordgo.MessageCreate) {
	createTimestamp, err := create.Timestamp.Parse()
	if err != nil {
		log.Printf("Error on parsing create timestamp: %v", err)
	}

	message := models.DiscordMemberMessage{
		MessageId:          create.ID,
		DiscordTextChannel: models.DiscordTextChannel{
			ChannelId: create.ChannelID,
		},
		DiscordMember:      models.DiscordMember{
			MemberId: create.Author.ID,
		},
		CreatedAt:          createTimestamp,
		EditedAt:           createTimestamp,
		IsActive:           true,
		MentionedRoles:     strings.Join(create.MentionRoles,","),
		Content:            create.Content,
		HasEmbedded:        len(create.Embeds) > 0,
	}

	_, err = a.discordRepository.InsertDiscordMemberMessage(message)
	if err != nil {
		log.Printf("Error on inserting member message: %v", err)
	}
}

func (a *Aybush) onLevel(session *discordgo.Session, create *discordgo.MessageCreate) {
	a.levelManager.OnMessage(create)
}

func (a *Aybush) onCommandReceived(session *discordgo.Session, messageCreate *discordgo.MessageCreate) {
	messageContent := messageCreate.Message.Content
	isCommand := strings.HasPrefix(messageContent, COMMAND_PREFIX)
	if !isCommand {
		return
	}

	commandArguments := strings.Split(strings.TrimPrefix(messageContent, COMMAND_PREFIX), " ")
	if len(commandArguments) < 1 {
		return
	}

	cmd, ok := a.commands[commandArguments[0]]
	if !ok {
		log.Printf("[AybushBot] Invalid command received. The command entered is %v", commandArguments[0])
		return
	}

	if len(commandArguments) == 2 {
		if strings.ToLower(commandArguments[1]) == HELP_ARG {
			helpMsg := fmt.Sprintf("<@%v>, %v", messageCreate.Author.ID, cmd.Usage())
			_, err := session.ChannelMessageSend(messageCreate.ChannelID, helpMsg)
			if err != nil {
				log.Printf("[AybushBot] Error on sending command usage message to the channel: %v", err)
			}
			return
		}
	}

	response, err := cmd.Execute(messageCreate.Message)
	if err != nil {
		log.Printf("[AybushBot] Error on executing the command: %v", err)
		return
	}

	if response != "" {
		_, err = session.ChannelMessageSend(messageCreate.ChannelID, response)
		if err != nil {
			log.Printf("[AybushBot] Error on sending command response to the channel: %v", err)
		}
	}
}

func (a *Aybush) onURLSend(session *discordgo.Session, messageCreate *discordgo.MessageCreate) {
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

	msg := fmt.Sprintf("> <@%v>, %v", messageCreate.Message.Author.ID, configuration.Manager.UrlRestriction.WarningMessage)
	_, err := session.ChannelMessageSend(messageCreate.ChannelID, msg)
	if err != nil {
		log.Printf("[AybushBot] Error on sending warning message to channel: %v", err)
	}

	err = session.ChannelMessageDelete(messageCreate.ChannelID, messageCreate.Message.ID)
	if err != nil {
		log.Printf("[AybushBot] Error on deleting a message which contains a URL: %v", err)
	}
}

func (a *Aybush) onSpamCheck(session *discordgo.Session, messageCreate *discordgo.MessageCreate) {
	a.antiSpam.OnMessage(messageCreate.Message)
}

func (a *Aybush) muteUserOnSpam(guildId string, memberId string, spamMessages []*discordgo.Message) {
	log.Printf("User %v muted in guild %v.", memberId, guildId)

	err := a.discordConnection.GuildMemberRoleAdd(guildId, memberId, configuration.Manager.Roles.MuteRole)
	if err != nil {
		log.Printf("[AybushBot] Error on adding muted role to member: %v", err)
	}

	lastChannelId := ""
	for _, msg := range spamMessages {
		lastChannelId = msg.ChannelID

		err = a.discordConnection.ChannelMessageDelete(msg.ChannelID, msg.ID)
		if err != nil {
			log.Printf("[AybushBot] Error on deleting a spam message: %v", err)
		}
	}

	muteDurationInMutes := configuration.Manager.AntiSpam.Mute.Duration / 60000

	// After time passes, mute role will be removed from the guild member.
	go func() {
		time.Sleep(time.Duration(configuration.Manager.AntiSpam.Mute.Duration) * time.Millisecond)
		err = a.discordConnection.GuildMemberRoleRemove(guildId, memberId, configuration.Manager.Roles.MuteRole)
		if err != nil {
			log.Printf("[AybushBot] Error on removing muted role from member: %v", err)
		}

		botLogMsg := fmt.Sprintf("> <@%v> kullanıcısının **spam** sebebiyle verilen `%v` dakikalık susturması kaldırıldı.", memberId,
			muteDurationInMutes)
		_, err = a.discordConnection.ChannelMessageSend(configuration.Manager.Channels.BotLog, botLogMsg)
		if err != nil {
			log.Printf("[AybushBot] Error on writing log to bot log channel: %v", err)
		}

	}()

	notificationMessageToGuildChannel := fmt.Sprintf(configuration.Manager.AntiSpam.Mute.ChannelMessage, memberId, muteDurationInMutes)
	_, err = a.discordConnection.ChannelMessageSend(lastChannelId, notificationMessageToGuildChannel)
	if err != nil {
		log.Printf("[AybushBot] Error on sending mute notification message to guild channel: %v", err)
	}

	dmChannel, err := a.discordConnection.UserChannelCreate(memberId)
	if err != nil {
		log.Printf("[AybushBot] Error on creating DM channel: %v", err)
		return
	}

	notificationMessageToDMChannel := fmt.Sprintf(configuration.Manager.AntiSpam.Mute.Message, muteDurationInMutes)
	_, err = a.discordConnection.ChannelMessageSend(dmChannel.ID, notificationMessageToDMChannel)
	if err != nil{
		log.Printf("[AybushBot] Error on sending mute notification message to DM channel: %v", err)
	}



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
			Value:  fmt.Sprintf("<@%v>", memberId),
			Inline: true,
		},
		{
			Name:   "Kanal",
			Value:  fmt.Sprintf("<#%v>", lastChannelId),
			Inline: true,
		},
		{
			Name:   "Uygulayan",
			Value:  fmt.Sprintf("<@%v>", configuration.Manager.BotUserId),
			Inline: true,
		},
		{
			Name: "Sebep",
			Value: "Spam",
			Inline: true,
		},
		{
			Name:   "Süre",
			Value:  fmt.Sprintf("%vm", muteDurationInMutes),
			Inline: false,
		},
	}
	botLogEmbedMsg.Footer = &discordgo.MessageEmbedFooter{
		Text:         time.Now().Format(time.Stamp),
	}

	_, err = a.discordConnection.ChannelMessageSendEmbed(configuration.Manager.Channels.BotLog, botLogEmbedMsg)
	if err != nil {
		log.Printf("[AybushBot] Error on writing log to bot log channel: %v", err)
	}
}