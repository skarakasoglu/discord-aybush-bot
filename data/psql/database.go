package psql

import (
	"github.com/skarakasoglu/discord-aybush-bot/data/models"
)

type Manager struct{
	host string
	port int
	user string
	password string
	dbName string
}

func (m *Manager) InsertMember(member models.Member) (bool, error) {
	panic("implement me")
}

func (m *Manager) GetMember(member models.Member) (models.Member, error) {
	panic("implement me")
}

func (m *Manager) DeleteMember(member models.Member) (bool, error) {
	panic("implement me")
}

func (m *Manager) GetAllMembers() ([]models.Member, error) {
	panic("implement me")
}

func (m *Manager) InsertMessage(message models.MemberMessage) (bool, error) {
	panic("implement me")
}

func (m *Manager) InsertRole(role models.Role) (bool, error) {
	panic("implement me")
}

func (m *Manager) GetRoles() ([]models.Role, error) {
	panic("implement me")
}

func (m *Manager) InsertLevel(level models.Level) (bool, error) {
	panic("implement me")
}

func (m *Manager) GetLevels() ([]models.Level, error) {
	panic("implement me")
}

func (m *Manager) InsertMemberLevel(level models.MemberLevel) (bool, error) {
	panic("implement me")
}

func (m *Manager) GetMemberLevel(member models.Member) (models.MemberLevel, error) {
	panic("implement me")
}

func (m *Manager) GetAllMemberLevels() ([]models.MemberLevel, error) {
	panic("implement me")
}

func NewManager(host string, port int, user string, password string, dbName string) *Manager {
	return &Manager{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		dbName:   dbName,
	}
}

