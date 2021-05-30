package models

import "time"

type DiscordMemberTimeBasedExperience struct {
	Id int
	DiscordMember
	EarnedExperiencePoints uint64
	EarnedTimestamp time.Time
	ExperienceTypeId int
}
