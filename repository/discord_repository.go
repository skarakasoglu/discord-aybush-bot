package repository

import "github.com/skarakasoglu/discord-aybush-bot/data/models"

type DiscordRepository interface{
	InsertDiscordAttachment(attachment models.DiscordAttachment) (bool, error)
	GetDiscordAttachmentById(attachmentId string) (models.DiscordAttachment, error)
	DeleteDiscordAttachmentById(attachmentId string) (bool error)

	InsertDiscordLevel(level models.DiscordLevel) (bool, error)
	UpdateDiscordLevel(level models.DiscordLevel) (bool, error)
	GetAllDiscordLevels() ([]models.DiscordLevel, error)
	DeleteDiscordLevel(level int) (bool error)

	InsertDiscordMember(member models.DiscordMember) (bool, error)
	UpdateDiscordMemberById(member models.DiscordMember) (bool, error)
	GetDiscordMemberById(memberId string) (models.DiscordMember, error)
	GetAllDiscordMembers() ([]models.DiscordMember, error)
	DeleteDiscordMemberById(memberId string) (bool, error)

	InsertDiscordMemberLevel(level models.DiscordMemberLevel) (bool, error)
	UpdateDiscordMemberLevelById(level models.DiscordMemberLevel) (bool, error)
	GetAllDiscordMemberLevels() ([]models.DiscordMemberLevel, error)
	DeleteDiscordMemberLevelById(memberLevelId int) (bool, error)

	InsertDiscordMemberMessage(message models.DiscordMemberMessage) (bool, error)
	DeleteDiscordMemberMessage(message models.DiscordMemberMessage) (bool, error)
	GetDiscordMemberMessagesByMemberId(memberId string) ([]models.DiscordMemberMessage, error)

	InsertDiscordRole(role models.DiscordRole) (bool, error)
	UpdateDiscordRoleById(role models.DiscordRole) (bool, error)
	GetAllDiscordRoles() ([]models.DiscordRole, error)
	DeleteDiscordRoleById(roleId int) (bool, error)

	InsertDiscordTextChannel(channel models.DiscordTextChannel) (bool, error)
	UpdateDiscordTextChannelById(channel models.DiscordTextChannel) (bool, error)
	GetAllDiscordTextChannels() ([]models.DiscordTextChannel, error)
	DeleteDiscordTextChannelById(channel models.DiscordTextChannel) (bool, error)
}
