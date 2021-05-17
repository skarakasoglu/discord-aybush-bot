package commands

import (
	"github.com/bwmarrin/discordgo"
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
	cmd.onRank(message.Author)
	return "", nil
}


func (cmd *rankCommand) Usage() string{
	return ""
}