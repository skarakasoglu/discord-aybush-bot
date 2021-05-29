package level

import (
	"github.com/skarakasoglu/discord-aybush-bot/configuration"
	_ "github.com/skarakasoglu/discord-aybush-bot/configuration"
	"log"
	"sort"
)

type SortedMemberLevelStatuses []*MemberLevelStatus

func (s SortedMemberLevelStatuses) Len() int { return len(s) }
func (s SortedMemberLevelStatuses) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
	s[i].Position, s[j].Position = i + 1, j + 1
}
func (s SortedMemberLevelStatuses) Less(i, j int) bool { return s[i].ExperiencePoints < s[j].ExperiencePoints }

func (m *Manager) sortMemberLevels() {
	m.orderedMemberLevelStatusMtx.Lock()
	sort.Sort(sort.Reverse(SortedMemberLevelStatuses(m.orderedMemberLevelStatuses)))
	m.detectStandingChanges()
	m.orderedMemberLevelStatusMtx.Unlock()
}

func (m *Manager) detectStandingChanges() {
	for i, member := range m.orderedMemberLevelStatuses {
		if i >= gradedMemberCount {
			break
		}

		if member.MemberId == m.gradedMembers[i].MemberId {
			continue
		} else {

			guildId := member.GuildId
			otherMember := m.gradedMembers[i]

			log.Printf("[AybushBot::LevelManager] Member position changed. FirstMemberId: %v, FirstMemberUsername: %v#%v, FirstMemberPosition: %v, " +
				"SecondMemberId: %v, SecondMemberUsername: %v#%v, SecondMemberPosition: %v", member.MemberId, member.Username, member.Discriminator, member.Position,
				otherMember.MemberId, otherMember.Username, otherMember.Discriminator, otherMember.Position)

			err := m.session.GuildMemberRoleRemove(guildId, otherMember.MemberId, configuration.Manager.Roles.GradedMembersRole)
			if err != nil {
				log.Printf("Error on removing member role: %v", err)
			}

			err = m.session.GuildMemberRoleRemove(guildId, otherMember.MemberId, rolePositions[i])
			if err != nil {
				log.Printf("Error on removing member role: %v", err)
			}

			hasRole := func(roles []string, roleId string) bool {
				for _, memberRole := range roles {
					if roleId == memberRole {
						return true
					}
				}

				return false
			}

			if !hasRole(member.Member.Roles, configuration.Manager.Roles.GradedMembersRole) {
				err = m.session.GuildMemberRoleAdd(guildId, otherMember.MemberId, configuration.Manager.Roles.GradedMembersRole)
				if err != nil {
					log.Printf("Error on adding member role: %v", err)
				}
			}

			err = m.session.GuildMemberRoleAdd(guildId, otherMember.MemberId, rolePositions[i])
			if err != nil {
				log.Printf("Error on adding member role: %v", err)
			}

			m.gradedMembers[i] = member
		}
	}

}