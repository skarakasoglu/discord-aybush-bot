package configuration

import (
	"github.com/spf13/viper"
	"log"
)

var (
	Manager manager
)

type manager struct{
	BotRoleId string
	BaseImagePath string
	Roles roles
	TwitchApi twitchApi
	Channels channels
	PresenceUpdate presenceUpdate
	Greeting greeting
	Ticket ticket
	UrlRestriction urlRestriction
	AntiSpam antiSpam
	LoveMeter loveMeter
	RockPaperScissors rockPaperScissors
}

func ReadConfigurationFile(path string, fileName string) {
	viper.SetConfigName(fileName)
	viper.AddConfigPath(path)

	if err := viper.ReadInConfig(); err !=  nil {
		log.Fatalf("Error reading configuration file: %v", err)
	}

	err := viper.Unmarshal(&Manager)
	if err != nil {
		log.Fatalf("Error decoding configuration file into struct: %v", err)
	}
}

type roles struct{
	DefaultMemberRole string
	MuteRole string
	ModerationRoles []string
}

type twitchApi struct{
	Address string
	Port int
}

type channels struct{
	BotLog string
	Aybus string
	Sohbet string
}

type greeting struct{
	GreetingChannel string
	GreetingDirectMessage string
	GreetingMessages []string
}

type presenceUpdate struct{
	PresenceUpdateFrequency int
	Statuses []string
}

type urlRestriction struct{
	WarningMessage string
	RestrictedChannels []string
}

type ticket struct{
	MessageId string
	Reaction string
	RoleId string
}

type antiSpam struct{
	Mute mute
	MaxInterval int
	MaxDuplicatesInterval int
}

type mute struct{
	Threshold int
	MaxDuplicates int
	Message string
	ChannelMessage string
	Duration int
}

type loveMeter struct {
	Texts []string
}

type rockPaperScissors struct{
	HostWins string
	AwayWins string
	Draw string
}