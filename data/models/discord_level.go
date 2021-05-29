package models

type DiscordLevel struct{
	Id int
	RequiredExperiencePoints uint64
	MaximumExperiencePoints uint64
	DiscordRole
}
