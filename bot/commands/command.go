package commands

import "github.com/bwmarrin/discordgo"

type Command interface{
	Name() string
	Execute(message *discordgo.Message) (string, error)
	ResponseMessage() string
	Usage() string
}