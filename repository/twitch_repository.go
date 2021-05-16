package repository

import "github.com/skarakasoglu/discord-aybush-bot/data/models"

type TwitchRepository interface{
	InsertTwitchBotAutoBroadcastMessage(messages models.TwitchBotAutoBroadcastMessage) (bool, error)
	UpdateTwitchBotAutoBroadcastMessageById(messages models.TwitchBotAutoBroadcastMessage) (bool, error)
	GetAllTwitchBotAutoBroadcastMessages() ([]models.TwitchBotAutoBroadcastMessage, error)
	DeleteTwitchBotAutoBroadcastMessageById(messageId int) (bool, error)

	InsertTwitchBotCommand(command models.TwitchBotCommand) (bool, error)
	UpdateTwitchBotCommandById(command models.TwitchBotCommand) (bool, error)
	GetAllTwitchBotCommands(command models.TwitchBotCommand) ([]models.TwitchBotCommand, error)
	DeleteTwitchBotCommandById(commandId int) (bool, error)

	InsertTwitchBotMessage(message models.TwitchBotMessage) (bool, error)
	UpdateTwitchBotMessageById(message models.TwitchBotMessage) (bool, error)
	GetAllTwitchBotMessages() ([]models.TwitchBotMessage, error)
	DeleteTwitchBotMessageById(messageId int) (bool, error)

	InsertTwitchBotMessageType(message models.TwitchBotMessageType) (bool, error)
	UpdateTwitchBotMessageTypeById(messageType models.TwitchBotMessageType) (bool, error)
	GetAllTwitchBotMessageTypes() ([]models.TwitchBotMessageType, error)
	DeleteTwitchBotMessageTypeById(messageTypeId int) (bool, error)

	InsertTwitchBotUserNoticeMessage(message models.TwitchBotUserNoticeMessage) (bool, error)
	UpdateTwitchBotUserNoticeMessageById(message models.TwitchBotUserNoticeMessage) (bool, error)
	GetAllTwitchBotUserNoticeMessages() ([]models.TwitchBotUserNoticeMessage, error)
	DeleteTwitchBotUserNoticeMessageById(messageId int) (bool, error)
}