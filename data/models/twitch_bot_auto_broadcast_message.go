package models

type TwitchBotAutoBroadcastMessage struct{
	Id string
	Message TwitchBotMessage
	IntervalSeconds int
}