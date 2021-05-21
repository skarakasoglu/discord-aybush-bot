package antispam

import "github.com/bwmarrin/discordgo"

type ProtectionConfig struct{
	Threshold int
	MaxDuplicates int
	Callback func(string, string, []*discordgo.Message)
}

type options struct{
	protectionConfigurations []ProtectionConfig

	//General variables
	maxInterval int
	maxDuplicatesInterval int
	exemptRoles []string
	ignoredUsers []string
	ignoredChannels []string
}