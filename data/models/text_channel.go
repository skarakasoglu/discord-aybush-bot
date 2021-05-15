package models

type TextChannel struct{
	Id int
	ChannelId string
	Name string
	IsNsfw bool
	CreatedAt int64
}
