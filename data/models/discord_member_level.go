package models

import "time"

type DiscordMemberLevel struct{
	Id int
	DiscordMember
	CurrentLevel DiscordLevel
	NextLevel DiscordLevel
	ExperiencePoints uint64
	LastMessageTimestamp time.Time
	MessageCount uint64
	ActiveVoiceMinutes uint64
}
