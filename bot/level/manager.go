package level

import (
	"github.com/bwmarrin/discordgo"
	"github.com/skarakasoglu/discord-aybush-bot/configuration"
	"github.com/skarakasoglu/discord-aybush-bot/data"
	"github.com/skarakasoglu/discord-aybush-bot/data/models"
	"log"
	"math/rand"
	"time"
)

type MemberLevelStatus struct{
	models.Member
	CurrentLevel models.Level
	NextLevel models.Level
	CurrentExperiencePoints int64
	LastMessageTimestamp int64
}

var (
	randSrc = rand.NewSource(time.Now().UnixNano())
	rnd = rand.New(randSrc)

	notSubNotBoosterTextMin = 1
	notSubNotBoosterTextMax = 15
	notSubNotBoosterVoiceMin = 1
	notSubNotBoosterVoiceMax = 15

	notBoosterButSubTextMin = 5
	notBoosterButSubTextMax = 15
	notBoosterButSubVoiceMin = 5
	notBoosterButSubVoiceMax = 15

	notSubButBoosterTextMin = 5
	notSubButBoosterTextMax = 15
	notSubButBoosterVoiceMin = 5
	notSubButBoosterVoiceMax = 15

	bothSubAndBoosterTextMin = 10
	bothSubAndBoosterTextMax = 15
	bothSubAndBoosterVoiceMin = 10
	bothSubAndBoosterVoiceMax = 15
)

type manager struct{
	levels []models.Level
	repository data.Repository
	memberLevelStatuses map[string]MemberLevelStatus
}

func NewManager(repository data.Repository) *manager{
	return &manager{
		repository: repository,
	}
}

func (m *manager) Start() {
	var err error
	m.levels, err = m.repository.GetLevels()
	if err != nil {
		log.Printf("Error on fetching levels: %v", err)
	}

	memberLevels, err := m.repository.GetAllMemberLevels()
	if err != nil {
		log.Printf("Error on fetching all member levels: %v", err)
	}

	for _, memberLevel := range memberLevels {
		var memberLevelStatus MemberLevelStatus
		memberLevelStatus.Member = memberLevel.Member
		memberLevelStatus.CurrentExperiencePoints = memberLevel.ExperiencePoints
		memberLevelStatus.LastMessageTimestamp = memberLevel.LastMessageTimestamp.Unix()

		for _, level := range m.levels {
			if memberLevelStatus.CurrentExperiencePoints > level.RequiredExperiencePoints {
				memberLevelStatus.CurrentLevel = level
				continue
			} else if memberLevelStatus.CurrentExperiencePoints < level.RequiredExperiencePoints {
				memberLevelStatus.NextLevel = level
				break
			}
		}

		m.memberLevelStatuses[memberLevelStatus.MemberId] = memberLevelStatus
	}
}

func (m *manager) OnMessage(session *discordgo.Session, messageCreate *discordgo.MessageCreate) {
	sentTime, err := messageCreate.Message.Timestamp.Parse()
	if err != nil {
		log.Printf("Error on parsing timestamp: %v", err)
		return
	}

	memberLevelStatus, ok := m.memberLevelStatuses[messageCreate.Member.User.ID]
	if !ok {
		memberLevelStatus = MemberLevelStatus{
			Member:                  models.Member{},
			CurrentLevel:            models.Level{
				Id: 0,
			},
			NextLevel:               m.levels[0],
			CurrentExperiencePoints: 0,
		}
	} else {
		timeDiff := sentTime.Unix() - memberLevelStatus.LastMessageTimestamp

		if timeDiff < 60 {
			return
		}
	}

	acquiredExperiencePoints := 0

	isSub, isBooster := false, false

	for _, roleId := range messageCreate.Member.Roles {
		if roleId == configuration.Manager.Roles.SubRole {
			isSub = true
		} else if roleId == configuration.Manager.Roles.ServerBoosterRole {
			isBooster = true
		}
	}

	if isSub {
		if isBooster {
			acquiredExperiencePoints = rnd.Intn(bothSubAndBoosterTextMax) + bothSubAndBoosterTextMin
		} else {
			acquiredExperiencePoints = rnd.Intn(notBoosterButSubTextMax) + notBoosterButSubTextMin
		}
	} else {
		if isBooster {
			acquiredExperiencePoints = rnd.Intn(notSubButBoosterTextMax) + notSubButBoosterTextMin
		} else {
			acquiredExperiencePoints = rnd.Intn(notSubNotBoosterTextMax) + notSubNotBoosterTextMin
		}
	}

	memberLevelStatus.LastMessageTimestamp = sentTime.Unix()
	memberLevelStatus.CurrentExperiencePoints += int64(acquiredExperiencePoints)

	if memberLevelStatus.CurrentExperiencePoints > memberLevelStatus.NextLevel.RequiredExperiencePoints && memberLevelStatus.NextLevel.Id < 100 {
		m.onLevelUp(session)
	} else {
		return
	}

	memberLevelStatus.CurrentLevel = memberLevelStatus.NextLevel

	if len(m.levels) > memberLevelStatus.CurrentLevel.Id + 1 {
		memberLevelStatus.NextLevel = m.levels[memberLevelStatus.CurrentLevel.Id + 1]
	} else {
		memberLevelStatus.NextLevel = models.Level{Id: 100}
	}

}

func (m *manager) onLevelUp(session *discordgo.Session) {
}