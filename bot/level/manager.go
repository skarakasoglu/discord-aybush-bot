package level

import (
	"bytes"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/skarakasoglu/discord-aybush-bot/configuration"
	"github.com/skarakasoglu/discord-aybush-bot/data/models"
	"github.com/skarakasoglu/discord-aybush-bot/repository"
	"image/png"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"
)

type ExpType int

const (
	ExpTypeVoice ExpType = iota
	ExpTypeText

	RoleEmojiLength = 5
)

var experienceTypes = map[ExpType]string{
	ExpTypeVoice: "Voice",
	ExpTypeText: "Text",
}

var expDemotionLevels = []int{50, 75}

var rolePositions []string

func (e ExpType) String() string {
	str, ok := experienceTypes[e]

	if ok {
		return str
	} else {
		return ""
	}
}

type MemberLevelStatus struct{
	CurrentLevel models.DiscordLevel
	NextLevel models.DiscordLevel
	models.DiscordMemberLevel
	Member *discordgo.Member
	Position int
}

type memberVoiceChanged struct{
	memberId string
	shouldEarn bool
}

const (
	textChannelEarningTimeoutSeconds = 60
	voiceChannelEarningTimeoutSeconds = 60

	expDemotionPercentage = 20.0

	notSubNotBoosterTextMin = 1
	notSubNotBoosterTextMax = 15
	notSubNotBoosterVoiceMin = 1
	notSubNotBoosterVoiceMax = 15

	notBoosterButSubTextMin = 5
	notBoosterButSubTextMax = 20
	notBoosterButSubVoiceMin = 5
	notBoosterButSubVoiceMax = 20

	notSubButBoosterTextMin = 5
	notSubButBoosterTextMax = 20
	notSubButBoosterVoiceMin = 5
	notSubButBoosterVoiceMax = 20

	bothSubAndBoosterTextMin = 10
	bothSubAndBoosterTextMax = 25
	bothSubAndBoosterVoiceMin = 10
	bothSubAndBoosterVoiceMax = 25

	gradedMemberCount = 3
)

var (
	randSrc = rand.NewSource(time.Now().UnixNano())
	rnd = rand.New(randSrc)
)

type ReloadMessage struct{
	ReloadMemberLevels bool
	ReloadDiscordMemberLevelMessages bool
	ReloadLevels bool
}

type Manager struct{
	running bool

	levelUpMessages []models.DiscordLevelUpMessage
	levels []models.DiscordLevel
	discordRepository repository.DiscordRepository
	memberLevelStatuses map[string]*MemberLevelStatus
	orderedMemberLevelStatuses []*MemberLevelStatus

	memberLevelStatusMtx sync.RWMutex
	orderedMemberLevelStatusMtx sync.RWMutex
	membersInVoiceMtx sync.Mutex

	ignoredTextChannels map[string]string
	ignoredVoiceChannels map[string]string

	session *discordgo.Session

	membersInVoice map[string]string

	reloadChan chan ReloadMessage
	onRankQueryChan chan *discordgo.User
	onVoiceChan chan memberVoiceChanged
	onMessageChan chan *discordgo.MessageCreate

	gradedMembers []*MemberLevelStatus
}

func NewManager(session *discordgo.Session, discordRepository repository.DiscordRepository, ignoredTextChannels []string, ignoredVoiceChannels []string) *Manager {
	ignoredTextChannelMap := make(map[string]string)
	for _, val := range ignoredTextChannels {
		ignoredTextChannelMap[val] = val
	}

	ignoredVoiceChannelMap := make(map[string]string)
	for _, val := range ignoredVoiceChannels {
		ignoredVoiceChannelMap[val] = val
	}

	rolePositions = []string{configuration.Manager.Roles.ServerFirstMemberRole, configuration.Manager.Roles.ServerSecondMemberRole, configuration.Manager.Roles.ServerThirdMemberRole}


	return &Manager{
		discordRepository: discordRepository,
		session: session,
		memberLevelStatuses: make(map[string]*MemberLevelStatus),
		membersInVoice: make(map[string]string),
		reloadChan: make(chan ReloadMessage, 500),
		onRankQueryChan: make(chan *discordgo.User, 500),
		onMessageChan: make(chan *discordgo.MessageCreate, 500),
		onVoiceChan: make(chan memberVoiceChanged, 500),
		gradedMembers: make([]*MemberLevelStatus, gradedMemberCount),
		ignoredTextChannels: ignoredTextChannelMap,
		ignoredVoiceChannels: ignoredVoiceChannelMap,
	}
}

func (m *Manager) Start() {
	m.running = true

	m.loadDiscordLevels()
	m.loadDiscordLevelUpMessages()
	m.loadDiscordMemberLevels()

	go m.giveExperienceActiveVoiceUsers()
	go m.workAsync()
}

func (m *Manager) Stop() {
	m.running = false
}

func (m *Manager) loadDiscordLevels() {
	var err error
	m.levels, err = m.discordRepository.GetAllDiscordLevels()
	if err != nil {
		log.Printf("[AybushBot::LevelManager] Error on fetching levels: %v", err)
	}
}

func (m *Manager) loadDiscordLevelUpMessages() {
	var err error
	m.levelUpMessages, err = m.discordRepository.GetAllDiscordLevelUpMessages()
	if err != nil {
		log.Printf("[AybushBot::LevelManager] Error on fetching level up messages: %v", err)
	}
}

func (m *Manager) loadDiscordMemberLevels() {
	memberLevels, err := m.discordRepository.GetAllDiscordMemberLevels()
	if err != nil {
		log.Printf("[AybushBot::LevelManager] Error on fetching all member levels: %v", err)
	}

	var wg sync.WaitGroup

	m.orderedMemberLevelStatusMtx.Lock()
	m.orderedMemberLevelStatuses = make([]*MemberLevelStatus, len(memberLevels))
	m.orderedMemberLevelStatusMtx.Unlock()

	log.Println("[AybushBot:::LevelManager] Discord member levels are being initialized.")
	for i, memberLevel := range memberLevels {
		wg.Add(1)
		go func(memberLevelParam models.DiscordMemberLevel, position int) {
			defer wg.Done()

			var memberLevelStatus MemberLevelStatus
			memberLevelStatus.DiscordMemberLevel = memberLevelParam

			member, err := m.session.GuildMember(memberLevelParam.GuildId, memberLevelParam.MemberId)
			if err != nil {
				log.Printf("[AybushBot::LevelManager] Error on obtaining member: %v", err)
				return
			}
			memberLevelStatus.Member = member
			memberLevelStatus.CurrentLevel = memberLevelParam.CurrentLevel
			memberLevelStatus.NextLevel = memberLevelParam.NextLevel

			userRole := m.levels[memberLevelStatus.CurrentLevel.Id].DiscordRole
			err = m.session.GuildMemberRoleAdd(memberLevelStatus.GuildId, memberLevelStatus.MemberId, userRole.RoleId)
			if err != nil {
				log.Printf("[AybushBot::LevelManager] Error on assigning role to member: %v", err)
			}

			m.memberLevelStatusMtx.Lock()
			m.memberLevelStatuses[memberLevelStatus.MemberId] = &memberLevelStatus
			m.memberLevelStatusMtx.Unlock()

			memberLevelStatus.Position = position + 1
			log.Printf("[AybushBot::LevelManager] MemberLevelStatusId: %v, MemberId: %v, GuildId: %v, Username: %v#%v, Position: %v, Exp: %v, CurrentLevel: %v, NextLevel: %v",
				memberLevelParam.Id, memberLevelParam.MemberId,
				memberLevelParam.GuildId, memberLevelParam.Username, memberLevelParam.Discriminator, memberLevelStatus.Position, memberLevelParam.ExperiencePoints,
				memberLevelParam.CurrentLevel, memberLevelParam.NextLevel)

			hasRole := func(roles []string, roleId string) bool {
				for _, memberRole := range roles {
					if roleId == memberRole {
						return true
					}
				}

				return false
			}
			for j, role := range rolePositions {
				if position == j {
					if !hasRole(member.Roles, role) {
						err = m.session.GuildMemberRoleAdd(memberLevelParam.GuildId, member.User.ID, role)
						if err != nil {
							log.Printf("Error on adding member role: %v", err)
						}
					}
				} else {
					if hasRole(member.Roles, role) {
						err = m.session.GuildMemberRoleRemove(memberLevelParam.GuildId, member.User.ID, role)
						if err != nil {
							log.Printf("Error on removing member role: %v", err)
						}
					}
				}
			}

			if position < gradedMemberCount {
				m.gradedMembers[position] = &memberLevelStatus

				if !hasRole(member.Roles, configuration.Manager.Roles.GradedMembersRole) {
					err = m.session.GuildMemberRoleAdd(memberLevelParam.GuildId, member.User.ID, configuration.Manager.Roles.GradedMembersRole)
					if err != nil {
						log.Printf("Error on adding member role: %v", err)
					}
				}
			} else {
				if hasRole(member.Roles, configuration.Manager.Roles.GradedMembersRole) {
					err = m.session.GuildMemberRoleRemove(memberLevelParam.GuildId, member.User.ID, configuration.Manager.Roles.GradedMembersRole)
					if err != nil {
						log.Printf("Error on removing member role: %v", err)
					}
				}
			}

			m.orderedMemberLevelStatusMtx.Lock()
			m.orderedMemberLevelStatuses[position] = &memberLevelStatus
			m.orderedMemberLevelStatusMtx.Unlock()
		}(memberLevel, i)
	}
	wg.Wait()

	log.Println("[AybushBot:::LevelManager] Discord member levels were initialized successfully.")
}

func (m *Manager) giveExperienceActiveVoiceUsers() {
	for m.running {
		m.membersInVoiceMtx.Lock()
		for _, memberId := range m.membersInVoice {
			m.earnExperienceFromVoice(memberId)
		}
		m.membersInVoiceMtx.Unlock()

		time.Sleep(time.Second * time.Duration(voiceChannelEarningTimeoutSeconds))
	}
}

func (m *Manager) workAsync() {
	for m.running {
		select {
		case user := <- m.onRankQueryChan:
			err := m.rankQueried(user)
			if err != nil {
				log.Printf("[AybushBot::LevelManager] Error on querying the rank of the user: %v", err)
			}
		case memberVoiceChange := <- m.onVoiceChan:
			m.membersInVoiceMtx.Lock()
			if memberVoiceChange.shouldEarn {
				m.membersInVoice[memberVoiceChange.memberId] = memberVoiceChange.memberId
			} else {
				delete(m.membersInVoice, memberVoiceChange.memberId)
			}
			m.membersInVoiceMtx.Unlock()
		case messageCreate := <- m.onMessageChan:
			m.earnExperienceFromMessage(messageCreate)
		case reloadMsg := <- m.reloadChan:
			log.Printf("[AybushBot::LevelManager] Reload message %+v", reloadMsg)

			if reloadMsg.ReloadLevels {
				m.loadDiscordLevels()
			}
			if reloadMsg.ReloadDiscordMemberLevelMessages {
				m.loadDiscordLevelUpMessages()
			}
			if reloadMsg.ReloadMemberLevels {
				m.loadDiscordMemberLevels()
			}
		}
	}
}

func (m *Manager) OnReload(reload ReloadMessage) {
	m.reloadChan <- reload
}

func (m *Manager) OnRankQuery(user *discordgo.User) {
	m.onRankQueryChan <- user
}

func (m *Manager) OnVoiceUpdate(update *discordgo.VoiceStateUpdate) {
	_, isIgnoredVoiceChannel := m.ignoredVoiceChannels[update.ChannelID]

	member, err := m.session.GuildMember(update.GuildID, update.UserID)
	if err != nil {
		log.Printf("[AybushBot::LevelManager] Error on obtaining guild member: %v", err)
		return
	}

	if member.User.Bot {
		return
	}

	if update.ChannelID != "" && !update.SelfDeaf && !update.Deaf && !isIgnoredVoiceChannel {
		voiceChanged := memberVoiceChanged{
			memberId:   update.UserID,
			shouldEarn: true,
		}
		log.Printf("[AybushBot::LevelManager] VoiceUpdate: MemberId: %v, ShouldEarn: %v, ChannelId: %v, Deaf: %v, SelfDeaf: %v", voiceChanged.memberId, voiceChanged.shouldEarn, update.ChannelID, update.Deaf, update.SelfDeaf)

		m.onVoiceChan <- voiceChanged
	} else if (update.BeforeUpdate != nil && update.ChannelID == "") || (update.ChannelID != "" && (update.SelfDeaf || update.Deaf)) || isIgnoredVoiceChannel {
		voiceChanged := memberVoiceChanged{
			memberId:   update.UserID,
			shouldEarn: false,
		}
		log.Printf("[AybushBot::LevelManager] VoiceUpdate: MemberId: %v, ShouldEarn: %v, ChannelId: %v, Deaf: %v, SelfDeaf: %v", voiceChanged.memberId, voiceChanged.shouldEarn, update.ChannelID, update.Deaf, update.SelfDeaf)

		m.onVoiceChan <- voiceChanged
	}
}

func (m *Manager) OnMessage(messageCreate *discordgo.MessageCreate) {
	m.onMessageChan <- messageCreate
}


func (m *Manager) earnExperienceFromMessage(messageCreate *discordgo.MessageCreate) {
	if messageCreate.Author.Bot {
		return
	}

	_, isIgnoredTextChannel := m.ignoredTextChannels[messageCreate.ChannelID]
	if isIgnoredTextChannel {
		return
	}

	sentTime, err := messageCreate.Message.Timestamp.Parse()
	if err != nil {
		log.Printf("[AybushBot::LevelManager] Error on parsing timestamp: %v", err)
		return
	}

	m.memberLevelStatusMtx.RLock()
	status, ok := m.memberLevelStatuses[messageCreate.Author.ID]
	m.memberLevelStatusMtx.RUnlock()

	if !ok {
		status, _ = m.createMemberLevel(messageCreate.Author.ID, sentTime)
		if status == nil {
			return
		}
	} else {
		timeDiff := sentTime.Unix() - status.LastMessageTimestamp.Unix()

		if timeDiff < textChannelEarningTimeoutSeconds {
			status.MessageCount++
			return
		}
	}

	status.MessageCount++
	status.LastMessageTimestamp = sentTime

	m.earnExperience(status, ExpTypeText)
}

func (m *Manager) earnExperienceFromVoice(memberId string) {
	m.memberLevelStatusMtx.RLock()
	status, ok := m.memberLevelStatuses[memberId]
	m.memberLevelStatusMtx.RUnlock()

	if !ok {
		status, _ = m.createMemberLevel(memberId, time.Unix(0, 0))
		if status == nil {
			return
		}
	}

	status.ActiveVoiceMinutes += voiceChannelEarningTimeoutSeconds / 60
	m.earnExperience(status, ExpTypeVoice)
}

func (m *Manager) earnExperience(status *MemberLevelStatus, expType ExpType) {
	earnedExperience := 0

	if expType == ExpTypeVoice {
		earnedExperience = m.calculateEarnedExperience(status, bothSubAndBoosterVoiceMax, bothSubAndBoosterVoiceMin,
			notBoosterButSubVoiceMax, notBoosterButSubVoiceMin, notSubButBoosterVoiceMax, notSubButBoosterVoiceMin, notSubNotBoosterVoiceMax, notSubNotBoosterVoiceMin)
	} else if expType == ExpTypeText {
		earnedExperience = m.calculateEarnedExperience(status,
			bothSubAndBoosterTextMax, bothSubAndBoosterTextMin,
			notBoosterButSubTextMax, notBoosterButSubTextMin,
			notSubButBoosterTextMax, notSubButBoosterTextMin, notSubNotBoosterTextMax, notSubNotBoosterTextMin)
	} else {
		log.Printf("[AybushBot::LevelManager] Invalid experience type. ExpType: %v, Status: %+v", expType, status)
		return
	}

	status.ExperiencePoints += int64(earnedExperience)
	if status.DiscordMemberLevel.ExperiencePoints >= status.NextLevel.RequiredExperiencePoints && status.NextLevel.Id < 99 {
		m.memberLeveledUp(status)
	}

	_, err := m.discordRepository.UpdateDiscordMemberLevelById(status.DiscordMemberLevel)
	if err != nil {
		log.Printf("[AybushBot::LevelManager] Error on updating discord member level: %v", err)
		return
	}

	m.sortMemberLevels()

	log.Printf("[AybushBot::LevelManager] ExperienceType: %v, MemberId: %v, Username: %v#%v, EarnedExp: %v, CurrentLevel: %v, Exp: %v, NextLevel: %v, RequiredExp: %v", expType.String(),
		status.MemberId, status.Member.User.Username, status.Member.User.Discriminator, earnedExperience, status.CurrentLevel.Id, status.ExperiencePoints, status.NextLevel.Id, status.NextLevel.RequiredExperiencePoints)
}

func (m *Manager) calculateEarnedExperience(member *MemberLevelStatus, bothSubAndBoosterMax int, bothSubAndBoosterMin int, notBoosterButSubMax int, notBoosterButSubMin int,
	notSubButBoosterMax int, notSubButBoosterMin int, notSubNotBoosterMax int, notSubNotBoosterMin int) int {
	earnedExperiencePoints := 0

	isSub, isBooster := false, false

	for _, roleId := range member.Member.Roles {
		if roleId == configuration.Manager.Roles.SubRole {
			isSub = true
		} else if roleId == configuration.Manager.Roles.ServerBoosterRole {
			isBooster = true
		}
	}

	min, max := 0, 0

	if isSub {
		if isBooster {
			max = bothSubAndBoosterMax
			min = bothSubAndBoosterMin
		} else {
			max = notBoosterButSubMax
			min = notBoosterButSubMin
		}
	} else {
		if isBooster {
			max = notSubButBoosterMax
			min = notSubButBoosterMin
		} else {
			max = notSubNotBoosterMax
			min = notSubNotBoosterMin
		}
	}

	for _, expDemotionLevel := range expDemotionLevels {
		if member.CurrentLevel.Id >= expDemotionLevel {
			max -= int(math.Round(float64(max) * expDemotionPercentage / 100))
			if min > 1 {
				min -= int(math.Round(float64(min) * expDemotionPercentage / 100.0))
			}
		}
	}

	log.Printf("[AybushBot::LevelManager] MemberId: %v, Username: %v#%v, Level:%v, MinExp: %v, MaxExp: %v, IsSub: %v, IsBooster: %v",
		member.MemberId, member.Username, member.Discriminator, member.CurrentLevel.Id, min, max, isSub, isBooster)
	earnedExperiencePoints = rnd.Intn(max - min + 1) + min
	return earnedExperiencePoints
}

func (m *Manager) memberLeveledUp(status *MemberLevelStatus) {
	if status.CurrentLevel.RoleId != status.NextLevel.RoleId {
		err := m.session.GuildMemberRoleRemove(status.DiscordMember.GuildId, status.DiscordMember.MemberId, status.CurrentLevel.RoleId)
		if err != nil {
			log.Printf("[AybushBot::LevelManager] Error on removing the current level role: %v", err)
		}

		err = m.session.GuildMemberRoleAdd(status.DiscordMember.GuildId, status.DiscordMember.MemberId, status.NextLevel.RoleId)
		if err != nil {
			log.Printf("[AybushBot::LevelManager] Error on assigning new level role to member: %v", err)
		}
	}

	status.CurrentLevel = status.NextLevel

	if len(m.levels) > status.CurrentLevel.Id + 1 {
		status.NextLevel = m.levels[status.CurrentLevel.Id + 1]
	} else {
		status.NextLevel = models.DiscordLevel{Id: 100}
	}

	index := rnd.Intn(len(m.levelUpMessages))
	levelUpMessage := fmt.Sprintf(m.levelUpMessages[index].Content, status.CurrentLevel.Id)

	message := fmt.Sprintf("<@%v>, %v", status.DiscordMember.MemberId, levelUpMessage)
	_, err := m.session.ChannelMessageSend(configuration.Manager.Channels.Aybus, message)
	if err != nil {
		log.Printf("[AybushBot::LevelManager] Error on sending message to channel: %v", err)
	}
}

func (m *Manager) createMemberLevel(memberId string, timestamp time.Time) (*MemberLevelStatus, error) {
	discordMember, err := m.discordRepository.GetDiscordMemberById(memberId)
	if err != nil {
		log.Printf("[AybushBot::LevelManager] Error on obtaining discord member id:%v : %v", memberId, err)
		return nil, err
	}

	member, err := m.session.GuildMember(discordMember.GuildId, memberId)
	if err != nil {
		log.Printf("[AybushBot::LevelManager] Error on obtaining member: %v", err)
		return nil, err
	}

	memberLevelStatus := &MemberLevelStatus{
		CurrentLevel: m.levels[0],
		NextLevel:               m.levels[1],
		DiscordMemberLevel: models.DiscordMemberLevel{
			DiscordMember:        discordMember,
			ExperiencePoints:     0,
			LastMessageTimestamp: timestamp,
		},
		Member: member,
	}

	lastInsertedId, err := m.discordRepository.InsertDiscordMemberLevel(memberLevelStatus.DiscordMemberLevel)
	if err != nil {
		log.Printf("[AybushBot::LevelManager] Error on inserting discord member level: %v", err)
		return nil, err
	}
	memberLevelStatus.DiscordMemberLevel.Id = lastInsertedId

	m.memberLevelStatusMtx.Lock()
	m.memberLevelStatuses[memberId] = memberLevelStatus
	m.memberLevelStatusMtx.Unlock()

	m.orderedMemberLevelStatusMtx.Lock()
	m.orderedMemberLevelStatuses = append(m.orderedMemberLevelStatuses, memberLevelStatus)
	m.orderedMemberLevelStatusMtx.Unlock()

	m.sortMemberLevels()

	return memberLevelStatus, nil
}

func (m *Manager) rankQueried(user *discordgo.User) error {
	var err error

	m.memberLevelStatusMtx.RLock()
	status, ok := m.memberLevelStatuses[user.ID]
	m.memberLevelStatusMtx.RUnlock()

	if !ok {
		status, err = m.createMemberLevel(user.ID, time.Unix(0, 0))
		if status == nil {
			return err
		}
	}

	avatar, err := m.session.UserAvatar(user.ID)
	if err != nil {
		log.Printf("[AybushBot::LevelManager] Error on decoding user avatar: %v", err)
		return err
	}

	imageWidth := 653.0
	imageHeight := 375.0

	roleName := "-"
	if len(status.CurrentLevel.DiscordRole.Name) > 5 {
		roleName = status.CurrentLevel.DiscordRole.Name[:len(status.CurrentLevel.DiscordRole.Name) - 5]
	}

	options := RankImageOptions{
		BackgroundImagePath: fmt.Sprintf("%v/%v", configuration.Manager.BaseImagePath,configuration.Manager.LevelSystem.Background),
		FontFace:            "fonts/Montserrat-Bold.ttf",
		Width:               int(imageWidth),
		Height:              int(imageHeight),
		AvatarWidth:         140,
		AvatarHeight:        140,
		AvatarArcX:          327,
		AvatarArcY:          83,
		AvatarRadius:        54,
		AvatarX:             260,
		AvatarY:             20,
		RankTextOptions: ShadowedTextOptions{
			Text:          fmt.Sprintf("#%v", status.Position),
			StrokeSize:    5,
			X:             625,
			Y:             30,
			Ax:            1,
			Ay:            0.5,
			ShadowOptions: ColorOptions{
				R:     0,
				G:     0,
				B:     0,
				Alpha: 5,
			},
			TextOptions:   ColorOptions{
				R:     255,
				G:     255,
				B:     255,
				Alpha: 255,
			},
		},
		UsernameTextOptions: ShadowedTextOptions{
			Text:          fmt.Sprintf("%v#%v", user.Username, user.Discriminator),
			StrokeSize:    6,
			X:             imageWidth / 2,
			Y:             170,
			Ax:            0.5,
			Ay:            0.5,
			ShadowOptions: ColorOptions{
				R:     0,
				G:     0,
				B:     0,
				Alpha: 5,
			},
			TextOptions:   ColorOptions{
				R:     255,
				G:     255,
				B:     255,
				Alpha: 255,
			},
		},
		RoleTextOptions: ShadowedTextOptions{
			Text:          fmt.Sprintf("%v", roleName),
			StrokeSize:    5,
			X:             imageWidth / 2,
			Y:             200,
			Ax:            0.5,
			Ay:            0.5,
			ShadowOptions: ColorOptions{
				R:     255,
				G:     255,
				B:     255,
				Alpha: 10,
			},
			TextOptions:   ColorOptions{
				R:     38,
				G:     39,
				B:     46,
				Alpha: 255,
			},
		},
		LevelTextOptions:    ShadowedTextOptions{
			Text:          fmt.Sprintf("Seviye %v", status.CurrentLevel.Id),
			StrokeSize:    5,
			X:             imageWidth / 2,
			Y:             260,
			Ax:            0.5,
			Ay:            0.5,
			ShadowOptions: ColorOptions{
				R:     0,
				G:     0,
				B:     0,
				Alpha: 5,
			},
			TextOptions:   ColorOptions{
				R:     255,
				G:     255,
				B:     255,
				Alpha: 255,
			},
		},
		ExpBarOptions:       ExpBarOptions{
			X:                  78,
			Y:                  300,
			Width:              500,
			Height:             17,
			Radius:             7,
			StrokeSize:         6,
			CurrentExperience:  status.ExperiencePoints,
			CurrentLevelRequiredExperience: status.CurrentLevel.RequiredExperiencePoints,
			RequiredExperience: status.NextLevel.RequiredExperiencePoints,
			ShadowOptions:      ColorOptions{
				R:     0,
				G:     0,
				B:     0,
				Alpha: 1,
			},
			EmptyBarOptions:    ColorOptions{
				R:     255,
				G:     255,
				B:     255,
				Alpha: 255,
			},
			FilledBarOptions:   ColorOptions{
				R:     139,
				G:     123,
				B:     255,
				Alpha: 255,
			},
		},
		CurrentExpOptions:   ShadowedTextOptions{
			Text:          fmt.Sprintf("%vxp / %vxp", status.ExperiencePoints, status.NextLevel.RequiredExperiencePoints),
			StrokeSize:    5,
			X:             575,
			Y:             285,
			Ax:            1,
			Ay:            0.5,
			ShadowOptions: ColorOptions{
				R:     0,
				G:     0,
				B:     0,
				Alpha: 5,
			},
			TextOptions:   ColorOptions{
				R:     255,
				G:     255,
				B:     255,
				Alpha: 255,
			},
		},
		CurrentLevelOptions: ShadowedTextOptions{
			Text:          fmt.Sprintf("%v", status.CurrentLevel.Id),
			StrokeSize:    5,
			X:             65,
			Y:             308,
			Ax:            1,
			Ay:            0.5,
			ShadowOptions: ColorOptions{
				R:     0,
				G:     0,
				B:     0,
				Alpha: 5,
			},
			TextOptions:   ColorOptions{
				R:     255,
				G:     255,
				B:     255,
				Alpha: 255,
			},
		},
		NextLevelOptions:    ShadowedTextOptions{
			Text:          fmt.Sprintf("%v", status.NextLevel.Id),
			StrokeSize:    5,
			X:             590,
			Y:             308,
			Ax:            0,
			Ay:            0.5,
			ShadowOptions: ColorOptions{
				R:     0,
				G:     0,
				B:     0,
				Alpha: 5,
			},
			TextOptions:   ColorOptions{
				R:     255,
				G:     255,
				B:     255,
				Alpha: 255,
			},
		},
		Avatar:              avatar,
	}

	image := createRankImage(options)
	imageBuffer := new(bytes.Buffer)
	err = png.Encode(imageBuffer, image)
	if err != nil {
		log.Printf("[AybushBot::LevelManager] Error on writing image to buffer: %v", err)
		return err
	}

	_, err = m.session.ChannelFileSend(configuration.Manager.Channels.Aybus, "rank.png", imageBuffer)
	if err != nil {
		log.Printf("[AybushBot::LevelManager] Error on sending file to channel: %v", err)
		return err
	}

	return nil
}