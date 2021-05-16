package models

type DiscordTextChannel struct{
	Id int
	ChannelId string
	Name string
	IsNsfw bool
	CreatedAt int64
}
