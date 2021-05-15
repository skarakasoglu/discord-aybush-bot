package data

import "github.com/skarakasoglu/discord-aybush-bot/data/models"

type Repository interface {
	InsertMember(member models.Member) (bool, error)
	GetMember(member models.Member) (models.Member, error)
	DeleteMember(member models.Member) (bool, error)
	GetAllMembers() ([]models.Member, error)

	InsertMessage(message models.MemberMessage) (bool, error)

	InsertRole(role models.Role) (bool, error)
	GetRoles() ([]models.Role, error)

	InsertLevel(level models.Level) (bool, error)
	GetLevels() ([]models.Level, error)

	InsertMemberLevel(level models.MemberLevel) (bool, error)
	GetMemberLevel(member models.Member) (models.MemberLevel, error)
	GetAllMemberLevels() ([]models.MemberLevel, error)
}