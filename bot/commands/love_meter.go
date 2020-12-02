package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
)

type loveMeterCommand struct{
	session *discordgo.Session
}

func NewLoveMeterCommand(session *discordgo.Session) Command{
	return &loveMeterCommand{
		session: session,
	}
}

func (cmd *loveMeterCommand) Name() string{
	return "aşk-ölçer"
}

func (cmd *loveMeterCommand) Execute(message *discordgo.Message) (string, error){
	if len(message.Mentions) == 1{
		member,err := cmd.session.GuildMember(message.GuildID, message.Mentions[0].ID)
		if err != nil{
			log.Printf("Error on optaining member: %v", err)
			return "", err
		}
		log.Printf("%v between %v",member.User.Username,message.Author.Username)
	}else{
		return fmt.Sprintf("<@%v>, %v", message.Author.ID, cmd.Usage()), nil
	}
	return  "", nil
}

func (cmd *loveMeterCommand) Usage() string{
	return "bu komutu\n\"\\t!aşk-ölçer <@kullanıcı-adı>\"\nşeklinde kullanabilirsiniz."
}
