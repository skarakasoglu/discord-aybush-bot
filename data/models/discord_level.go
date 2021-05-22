package models

type DiscordLevel struct{
	Id int
	RequiredExperiencePoints int64
	MaximumExperiencePoints int64
	DiscordRole
}
