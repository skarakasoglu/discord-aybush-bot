package commands

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

type rankCommand struct{
	onRank func(member *discordgo.User)
}

func NewRankCommand(onRank func(member *discordgo.User)) *rankCommand{
	return &rankCommand{onRank: onRank}
}

func (cmd *rankCommand) Name() string{
	return "rank"
}

func (cmd *rankCommand) Execute(message *discordgo.Message) (string, error) {
	log.Printf("[AybushBot::Rank] User queried rank. MemberId: %v, Username: %v#%v", message.Author.ID, message.Author.Username, message.Author.Discriminator)
	cmd.onRank(message.Author)
	return "", nil
}


func (cmd *rankCommand) Usage() string{
	return ""
}