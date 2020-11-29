package configuration

import (
	"github.com/spf13/viper"
	"log"
)

var (
	Manager manager
)

type manager struct{
	Roles roles
	Channels channels
	PresenceUpdate presenceUpdate
	Greeting greeting
	Ticket ticket
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
}

type channels struct{
	BotLog string
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

type ticket struct{
	MessageId string
	Reaction string
	RoleId string
}