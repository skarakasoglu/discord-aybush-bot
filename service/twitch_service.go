package service

import (
	"database/sql"
	"github.com/skarakasoglu/discord-aybush-bot/data/models"
	"log"
)

type TwitchService struct {
	db *sql.DB
}

func (t TwitchService) InsertTwitchBotAutoBroadcastMessage(messages models.TwitchBotAutoBroadcastMessage) (bool, error) {
	panic("implement me")
}

func (t TwitchService) UpdateTwitchBotAutoBroadcastMessageById(messages models.TwitchBotAutoBroadcastMessage) (bool, error) {
	panic("implement me")
}

func (t TwitchService) GetAllTwitchBotAutoBroadcastMessages() ([]models.TwitchBotAutoBroadcastMessage, error) {
	var autoBroadcastMessages []models.TwitchBotAutoBroadcastMessage

	query := `
		SELECT tbabm.id, tbabm.interval_seconds, tbm.id, tbm.content, tbm.minimum_bits, tbmt.id, tbmt.name
		FROM "twitch_bot_auto_broadcast_messages" as tbabm 
		join "twitch_bot_messages" as tbm on tbabm.twitch_bot_message_id = tbm.id
		JOIN "twitch_bot_message_types" as tbmt on tbm.twitch_bot_message_type_id = tbmt.id;
		`

	rows, err := t.db.Query(query)
	if err != nil {
		log.Printf("[TwitchService] Error on executing the query: %v", err)
		return autoBroadcastMessages, err
	}

	for rows.Next() {
		var autoBroadcastMessage models.TwitchBotAutoBroadcastMessage
		var message models.TwitchBotMessage
		var messageType models.TwitchBotMessageType

		err = rows.Scan(&autoBroadcastMessage.Id, &autoBroadcastMessage.IntervalSeconds, &message.Id, &message.Content, &message.MinimumBits, &messageType.Id, &messageType.Name)
		if err != nil {
			log.Printf("[TwitchService] Error on scanning the rows: %v", err)
			return autoBroadcastMessages, err
		}
		autoBroadcastMessage.Message = message
		autoBroadcastMessage.Message.Type = messageType

		autoBroadcastMessages = append(autoBroadcastMessages, autoBroadcastMessage)
	}

	return autoBroadcastMessages, nil
}

func (t TwitchService) DeleteTwitchBotAutoBroadcastMessageById(messageId int) (bool, error) {
	panic("implement me")
}

func (t TwitchService) InsertTwitchBotCommand(command models.TwitchBotCommand) (bool, error) {
	panic("implement me")
}

func (t TwitchService) UpdateTwitchBotCommandById(command models.TwitchBotCommand) (bool, error) {
	panic("implement me")
}

func (t TwitchService) GetAllTwitchBotCommands() ([]models.TwitchBotCommand, error) {
	var commands []models.TwitchBotCommand

	query := `
	select tbc.id, tbc.command, tbm.id, tbm.content, tbm.minimum_bits, tbmt.id, tbmt.name 
	from "twitch_bot_commands" as tbc join "twitch_bot_messages" as tbm on tbc.twitch_bot_message_id = tbm.id
	join "twitch_bot_message_types" as tbmt on tbmt.id = tbm.twitch_bot_message_type_id;
	`

	rows, err := t.db.Query(query)
	if err != nil {
		log.Printf("[TwitchService] Error on executing the query: %v", err)
		return commands, err
	}

	for rows.Next() {
		var command models.TwitchBotCommand
		var message models.TwitchBotMessage
		var messageType models.TwitchBotMessageType

		err = rows.Scan(&command.Id, &command.Command, &message.Id, &message.Content, &message.MinimumBits, &messageType.Id, &messageType.Name)
		if err != nil {
			log.Printf("[TwitchService] Error on scanning the rows: %v", err)
			return commands, err
		}
		command.Message = message
		command.Message.Type = messageType

		commands = append(commands, command)
	}

	return commands, nil
}

func (t TwitchService) DeleteTwitchBotCommandById(commandId int) (bool, error) {
	panic("implement me")
}

func (t TwitchService) InsertTwitchBotMessage(message models.TwitchBotMessage) (bool, error) {
	panic("implement me")
}

func (t TwitchService) UpdateTwitchBotMessageById(message models.TwitchBotMessage) (bool, error) {
	panic("implement me")
}

func (t TwitchService) GetAllTwitchBotMessages() ([]models.TwitchBotMessage, error) {
	var botMessages []models.TwitchBotMessage

	query := `SELECT tbm.id, tbm.content, tbm.minimum_bits, tbmt.id, tbmt.name 
				FROM "twitch_bot_messages" as tbm join "twitch_bot_message_types" as tbmt on tbm.twitch_bot_message_type_id = tbmt.id;`

	rows, err := t.db.Query(query)
	if err != nil {
		log.Printf("[TwitchService] Error on executing the query: %v", err)
		return botMessages, err
	}

	for rows.Next() {
		var message models.TwitchBotMessage
		var messageType models.TwitchBotMessageType

		err = rows.Scan(&message.Id, &message.Content, &message.MinimumBits, &messageType.Id, &messageType.Name)
		if err != nil {
			log.Printf("[TwitchService] Error on scanning the rows: %v", err)
			return botMessages, err
		}
		message.Type = messageType

		botMessages = append(botMessages, message)
	}

	return botMessages, nil
}

func (t TwitchService) DeleteTwitchBotMessageById(messageId int) (bool, error) {
	panic("implement me")
}

func (t TwitchService) InsertTwitchBotMessageType(message models.TwitchBotMessageType) (bool, error) {
	panic("implement me")
}

func (t TwitchService) UpdateTwitchBotMessageTypeById(messageType models.TwitchBotMessageType) (bool, error) {
	panic("implement me")
}

func (t TwitchService) GetAllTwitchBotMessageTypes() ([]models.TwitchBotMessageType, error) {
	panic("implement me")
}

func (t TwitchService) DeleteTwitchBotMessageTypeById(messageTypeId int) (bool, error) {
	panic("implement me")
}

func (t TwitchService) InsertTwitchBotUserNoticeMessage(message models.TwitchBotUserNoticeMessage) (bool, error) {
	panic("implement me")
}

func (t TwitchService) UpdateTwitchBotUserNoticeMessageById(message models.TwitchBotUserNoticeMessage) (bool, error) {
	panic("implement me")
}

func (t TwitchService) GetAllTwitchBotUserNoticeMessages() ([]models.TwitchBotUserNoticeMessage, error) {
	var noticeMessages []models.TwitchBotUserNoticeMessage

	query := `SELECT * FROM "twitch_bot_user_notice_messages";`

	rows, err := t.db.Query(query)
	if err != nil {
		log.Printf("[TwitchService] Error on executing the query: %v", err)
		return noticeMessages, err
	}

	for rows.Next() {
		var message models.TwitchBotUserNoticeMessage

		err = rows.Scan(&message.Id, &message.NoticeEvent, &message.Tier, &message.IsRecipientMe, &message.Content)
		if err != nil {
			log.Printf("[TwitchService] Error on scanning the rows: %v", err)
			return noticeMessages, err
		}

		noticeMessages = append(noticeMessages, message)
	}

	return noticeMessages, nil
}

func (t TwitchService) DeleteTwitchBotUserNoticeMessageById(messageId int) (bool, error) {
	panic("implement me")
}

func NewTwitchService(db *sql.DB) *TwitchService{
	return &TwitchService{db: db}
}
