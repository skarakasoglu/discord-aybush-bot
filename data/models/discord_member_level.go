package models

import "time"

type DiscordMemberLevel struct{
	Id int
	DiscordMember
	CurrentLevel DiscordLevel
	NextLevel DiscordLevel
	ExperiencePoints int64
	LastMessageTimestamp time.Time
	MessageCount int64
	ActiveVoiceMinutes int64
}
