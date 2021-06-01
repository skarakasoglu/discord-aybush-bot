package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/skarakasoglu/discord-aybush-bot/configuration"
	"log"
	"strings"
)

type joiningDateCommand struct{
	session *discordgo.Session
}

func NewJoiningDateCommand(session *discordgo.Session) Command{
	return &joiningDateCommand{
		session: session,
	}
}

func (cmd *joiningDateCommand) Name() string{
	return "katılma-tarihi"
}

func (cmd *joiningDateCommand) Execute(message *discordgo.Message) (string, error) {
	arguments := strings.Split(message.Content, " ")[1:]

	if len(arguments) < 1 {
		joinedAt, err := message.Member.JoinedAt.Parse()
		if err != nil {
			log.Printf("[AybushBot] Error on parsing timestamp data: %v", err)
			return "", err
		}

		response := fmt.Sprintf("> <@%v>, sunucuya katılma tarihin **%v**", message.Author.ID, joinedAt.Format("02 Jan 06 15:04"))
		return response, nil
	}

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
		return fmt.Sprintf("> <@%v>, bu komutu kullanmaya **yetkiniz** bulunmamaktadır.", message.Author.ID), nil
	}

	if len(message.Mentions) < 1 || len(message.Mentions) > 1 {
		return fmt.Sprintf("<@%v>, %v", message.Author.ID, cmd.Usage()), nil
	}

	member, err := cmd.session.GuildMember(message.GuildID, message.Mentions[0].ID)
	if err != nil {
		log.Printf("[AybushBot] Error on obtaining guild member: %v", err)
		return "", err
	}

	joinedAt, err := member.JoinedAt.Parse()
	if err != nil {
		log.Printf("[AybushBot] Error on parsing timestamp data: %v", err)
		return "", err
	}

	return fmt.Sprintf(cmd.ResponseMessage(), message.Author.ID, member.User.ID, joinedAt.Format("02 Jan 06 15:04")), nil
}

func (cmd *joiningDateCommand) ResponseMessage() string{
	return "> <@%v>, <@%v> kullanıcısının katılma tarihi **%v**"
}

func (cmd *joiningDateCommand) Usage() string{
	usageType1 := "\t>!katilma-tarihi"
	usageType2 := "\t>!katilma-tarihi `<kullanici-adi>` *(Moderasyon yetkisi gerektirir)*"
	return fmt.Sprintf("**bu komutu**\n%v\nşekillerinde kullanabilirsiniz.", strings.Join([]string{usageType1, usageType2}, "\n"))
}