package models

import "time"

type DiscordMemberMessage struct{
	Id int
	MessageId string
	DiscordTextChannel
	DiscordMember
	CreatedAt time.Time
	EditedAt time.Time
	IsActive bool
	MentionedRoles string
	Content string
	HasEmbedded bool
}
