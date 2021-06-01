package commands

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

type leaderboardCommand struct{
	onRank func(member *discordgo.User)
}

func NewLeaderboardCommand() *leaderboardCommand{
	return &leaderboardCommand{}
}

func (cmd *leaderboardCommand) Name() string{
	return "leaderboard"
}

func (cmd *leaderboardCommand) Execute(message *discordgo.Message) (string, error) {
	log.Printf("[AybushBot::Leaderboard] Leaderboard command received. MemberId: %v, Username: %v#%v", message.Author.ID, message.Author.Username, message.Author.Discriminator)
	return cmd.ResponseMessage(), nil
}

func (cmd* leaderboardCommand) ResponseMessage() string{
	return "Liderlik tablosuna https://leaderboard.aybushbot.com adresinden ulaşabilirsin."
}

func (cmd *leaderboardCommand) Usage() string{
	return ""
}