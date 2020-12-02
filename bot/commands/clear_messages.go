package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/skarakasoglu/discord-aybush-bot/configuration"
	"log"
	"strconv"
	"strings"
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
		log.Printf("Invalid clear command argument, expected integer: %v", err)
		return fmt.Sprintf("<@%v>, %v", message.Author.ID, cmd.Usage()) ,nil
	}

	if messageCount >= 99 {
		messageCount = 100
	}

	messages, err := cmd.session.ChannelMessages(message.ChannelID, messageCount, "", "", "")
	if err != nil {
		log.Printf("Error on fetching channel messages: %v", err)
		return "", err
	}

	var messageIds []string

	for _, msg := range messages {
		messageIds = append(messageIds, msg.ID)
	}

	err = cmd.session.ChannelMessagesBulkDelete(message.ChannelID, messageIds)
	if err != nil {
		log.Printf("Error on bulk deleting messages: %v", err)
		return "", err
	}
	log.Printf("%v#%v deleted %v messages in channel %v.", message.Author.Username, message.Author.Discriminator, messageCount, message.ChannelID)

	botLogMsg := fmt.Sprintf("> <@%v>, <#%v> kanalında **%v** adet mesaj temizledi.", message.Author.ID, message.ChannelID, messageCount)
	_, err = cmd.session.ChannelMessageSend(configuration.Manager.Channels.BotLog, botLogMsg)
	if err != nil {
		log.Printf("Error on writing lot go bot log channel: %v", err)
	}

	return fmt.Sprintf("%v adet mesaj temizlendi.", messageCount), nil
}

func (cmd *clearMessagesCommand) Usage() string{
	return "**bu komut,** \n> !temizle `<mesaj-sayısı(max: 99)`>\n şeklinde kullanılır. *(Moderasyon yetkisi gerektirir)*"
}