package models

type TwitchBotMessage struct{
	Id int
	Content string
	Type TwitchBotMessageType
	MinimumBits int
}
