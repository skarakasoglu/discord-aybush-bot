package models

type MemberMessage struct{
	Id int
	MessageId string
	TextChannel
	Member
	CreatedAt int64
	EditedAt int64
	IsActive bool
	MentionedRoles string
	Content string
	HasEmbedded bool
}
