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

type clearMessagesCommand struct{
	session *discordgo.Session
}

func NewClearMessageCommand(session *discordgo.Session) Command{
	return &clearMessagesCommand{
		session: session,
	}
}

func (cmd *clearMessagesCommand) Name() string{
	return "temizle"
}

func (cmd *clearMessagesCommand) Execute(message *discordgo.Message) (string, error) {
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
		return fmt.Sprintf("> <@%v>, bu komutu kullanmaya yetkiniz bulunmamaktadır.", message.Author.ID), nil
	}

	if len(arguments) < 1 {
		return fmt.Sprintf("<@%v>, %v", message.Author.ID, cmd.Usage()), nil
	}

	messageCount, err := strconv.Atoi(arguments[0])
	if err != nil {
		log.Printf("[AybushBot::ClearMessages] Invalid clear command argument, expected integer: %v", err)
		return fmt.Sprintf("<@%v>, %v", message.Author.ID, cmd.Usage()) ,nil
	}

	if messageCount >= 99 {
		messageCount = 100
	}

	messages, err := cmd.session.ChannelMessages(message.ChannelID, messageCount, "", "", "")
	if err != nil {
		log.Printf("[AybushBot::ClearMessages] Error on fetching channel messages: %v", err)
		return "", err
	}

	var messageIds []string

	for _, msg := range messages {
		messageIds = append(messageIds, msg.ID)
	}

	err = cmd.session.ChannelMessagesBulkDelete(message.ChannelID, messageIds)
	if err != nil {
		log.Printf("[AybushBot::ClearMessages] Error on bulk deleting messages: %v", err)
		return "", err
	}
	log.Printf("[AybushBot::ClearMessages] %v#%v deleted %v messages in channel %v.", message.Author.Username, message.Author.Discriminator, messageCount, message.ChannelID)

	embedBotLogMsg := embed.NewGenericEmbed("Moderasyon İşlemi", "")
	embedBotLogMsg.Color = 0xF97100
	embedBotLogMsg.Fields = []*discordgo.MessageEmbedField{
		{
			Name:   "İşlem",
			Value:  "Temizle",
			Inline: true,
		},
		{
			Name:   "Kanal",
			Value:  fmt.Sprintf("<#%v>", message.ChannelID),
			Inline: true,
		},
		{
			Name: "Uygulayan",
			Value: fmt.Sprintf("<@%v>", message.Author.ID),
			Inline: false,
		},
		{
			Name: "Adet",
			Value: fmt.Sprintf("%v", messageCount),
			Inline: false,
		},
	}
	embedBotLogMsg.Footer = &discordgo.MessageEmbedFooter{
		Text:         time.Now().Format(time.Stamp),
	}

	_, err = cmd.session.ChannelMessageSendEmbed(configuration.Manager.Channels.BotLog, embedBotLogMsg)
	if err != nil {
		log.Printf("[AybushBot::ClearMessages] Error on writing lot go bot log channel: %v", err)
	}

	return fmt.Sprintf("%v adet mesaj temizlendi.", messageCount), nil
}

func (cmd *clearMessagesCommand) Usage() string{
	return "**bu komut,** \n> !temizle `<mesaj-sayısı(max: 99)`>\n şeklinde kullanılır. *(Moderasyon yetkisi gerektirir)*"
}