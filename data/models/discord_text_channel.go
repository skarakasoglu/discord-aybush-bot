package models

import "time"

type DiscordTextChannel struct{
	Id int
	ChannelId string
	Name string
	IsNsfw bool
	CreatedAt time.Time
}
