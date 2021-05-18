package models

type DiscordMemberMessage struct{
	Id int
	MessageId string
	DiscordTextChannel
	DiscordMember
	CreatedAt int64
	EditedAt int64
	IsActive bool
	MentionedRoles string
	Content string
	HasEmbedded bool
}
