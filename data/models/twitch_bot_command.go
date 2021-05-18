package models

type TwitchBotCommand struct{
	Id int
	Command string
	Message TwitchBotMessage
}