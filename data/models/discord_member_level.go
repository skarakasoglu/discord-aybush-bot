package models

import "time"

type DiscordMemberLevel struct{
	Id int
	DiscordMember
	ExperiencePoints int64
	LastMessageTimestamp time.Time
}
