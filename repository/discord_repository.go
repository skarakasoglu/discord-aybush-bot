package repository

import "github.com/skarakasoglu/discord-aybush-bot/data/models"

type DiscordRepository interface{
	InsertDiscordAttachment(attachment models.DiscordAttachment) (int, error)
	GetDiscordAttachmentById(attachmentId string) (models.DiscordAttachment, error)
	DeleteDiscordAttachmentById(attachmentId string) (bool error)

	InsertDiscordLevel(level models.DiscordLevel) (int, error)
	UpdateDiscordLevel(level models.DiscordLevel) (bool, error)
	GetAllDiscordLevels() ([]models.DiscordLevel, error)
	DeleteDiscordLevel(level int) (bool error)

	InsertDiscordMember(member models.DiscordMember) (int, error)
	UpdateDiscordMemberById(member models.DiscordMember) (bool, error)
	GetDiscordMemberById(memberId string) (models.DiscordMember, error)
	GetAllDiscordMembers() ([]models.DiscordMember, error)
	DeleteDiscordMemberById(memberId string) (bool, error)

	InsertDiscordMemberLevel(level models.DiscordMemberLevel) (int, error)
	UpdateDiscordMemberLevelById(level models.DiscordMemberLevel) (bool, error)
	GetAllDiscordMemberLevels() ([]models.DiscordMemberLevel, error)
	DeleteDiscordMemberLevelById(memberLevelId int) (bool, error)

	InsertDiscordMemberMessage(message models.DiscordMemberMessage) (int, error)
	DeleteDiscordMemberMessage(message models.DiscordMemberMessage) (bool, error)
	GetDiscordMemberMessagesByMemberId(memberId string) ([]models.DiscordMemberMessage, error)

	InsertDiscordRole(role models.DiscordRole) (int, error)
	UpdateDiscordRoleById(role models.DiscordRole) (bool, error)
	GetAllDiscordRoles() ([]models.DiscordRole, error)
	DeleteDiscordRoleById(roleId string) (bool, error)

	InsertDiscordTextChannel(channel models.DiscordTextChannel) (int, error)
	UpdateDiscordTextChannelById(channel models.DiscordTextChannel) (bool, error)
	GetAllDiscordTextChannels() ([]models.DiscordTextChannel, error)
	DeleteDiscordTextChannelById(channelId string) (bool, error)

	InsertDiscordLevelUpMessage(message models.DiscordLevelUpMessage) (int, error)
	UpdateDiscordLevelUpMessage(message models.DiscordLevelUpMessage) (bool, error)
	GetAllDiscordLevelUpMessages() ([]models.DiscordLevelUpMessage, error)
	DeleteDiscordLevelUpMessageById(id int) (bool, error)

	InsertDiscordMemberTimeBasedExperience(experience models.DiscordMemberTimeBasedExperience) (int, error)

	InsertDiscordEpisodeExperiences(experience models.DiscordEpisodeExperience) (int, error)
	UpdateActiveDiscordEpisodeExperiences(experience models.DiscordEpisodeExperience) (bool, error)
	GetAllEpisodeExperiences() ([]models.DiscordEpisodeExperience, error)
}
