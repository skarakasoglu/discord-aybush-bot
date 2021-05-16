package models

type TwitchBotUserNoticeMessage struct{
	Id int
	NoticeEvent string
	Tier string
	IsRecipientMe bool
	Content string
}
