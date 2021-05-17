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
	"math/rand"
	"sort"
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
}

type SortedMemberLevelStatuses []*MemberLevelStatus

func (s SortedMemberLevelStatuses) Len() int { return len(s) }
func (s SortedMemberLevelStatuses) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s SortedMemberLevelStatuses) Less(i, j int) bool { return s[i].ExperiencePoints < s[j].ExperiencePoints }

type memberVoiceChanged struct{
	memberId string
	shouldEarn bool
}

const (
	textChannelEarningTimeoutSeconds = 60
	voiceChannelEarningTimeoutSeconds = 60

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

var (
	randSrc = rand.NewSource(time.Now().UnixNano())
	rnd = rand.New(randSrc)
)

type Manager struct{
	running bool

	levelUpMessages []models.DiscordLevelUpMessage
	levels []models.DiscordLevel
	discordRepository repository.DiscordRepository
	memberLevelStatuses map[string]*MemberLevelStatus
	orderedMemberLevelStatuses []*MemberLevelStatus

	memberLevelStatusMtx sync.RWMutex
	orderedMemberLevelStatusMtx sync.RWMutex

	ignoredTextChannels map[string]string
	ignoredVoiceChannels map[string]string

	session *discordgo.Session

	membersInVoice map[string]string

	onRankQueryChan chan *discordgo.User
	onVoiceChan chan memberVoiceChanged
	onMessageChan chan *discordgo.MessageCreate
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


	return &Manager{
		discordRepository: discordRepository,
		session: session,
		memberLevelStatuses: make(map[string]*MemberLevelStatus),
		membersInVoice: make(map[string]string),
		onRankQueryChan: make(chan *discordgo.User, 500),
		onMessageChan: make(chan *discordgo.MessageCreate, 500),
		onVoiceChan: make(chan memberVoiceChanged, 500),
		ignoredTextChannels: ignoredTextChannelMap,
		ignoredVoiceChannels: ignoredVoiceChannelMap,
	}
}

func (m *Manager) Start() {
	m.running = true

	var err error
	m.levels, err = m.discordRepository.GetAllDiscordLevels()
	if err != nil {
		log.Printf("[AybushBot::LevelManager] Error on fetching levels: %v", err)
	}

	m.levelUpMessages, err = m.discordRepository.GetAllDiscordLevelUpMessages()
	if err != nil {
		log.Printf("[AybushBot::LevelManager] Error on fetching level up messages: %v", err)
	}

	memberLevels, err := m.discordRepository.GetAllDiscordMemberLevels()
	if err != nil {
		log.Printf("[AybushBot::LevelManager] Error on fetching all member levels: %v", err)
	}

	m.memberLevelStatusMtx.Lock()
	m.orderedMemberLevelStatusMtx.Lock()

	defer m.memberLevelStatusMtx.Unlock()
	defer m.orderedMemberLevelStatusMtx.Unlock()

	for _, memberLevel := range memberLevels {
		log.Printf("[AybushBot::LevelManager] MemberLevelStatusId: %v, MemberId: %v, GuildId: %v, Username: %v#%v, Exp: %v", memberLevel.Id, memberLevel.MemberId,
			memberLevel.GuildId, memberLevel.Username, memberLevel.Discriminator, memberLevel.ExperiencePoints)
		var memberLevelStatus MemberLevelStatus
		memberLevelStatus.DiscordMemberLevel = memberLevel

		member, err := m.session.GuildMember(memberLevel.GuildId, memberLevel.MemberId)
		if err != nil {
			log.Printf("[AybushBot::LevelManager] Error on obtaining member: %v", err)
			continue
		}
		memberLevelStatus.Member = member

		for _, level := range m.levels {
			if memberLevelStatus.DiscordMemberLevel.ExperiencePoints > level.RequiredExperiencePoints {
				memberLevelStatus.CurrentLevel = level
				continue
			} else if memberLevelStatus.DiscordMemberLevel.ExperiencePoints < level.RequiredExperiencePoints {
				memberLevelStatus.NextLevel = level
				break
			}
		}

		userRole := m.levels[memberLevelStatus.CurrentLevel.Id].DiscordRole
		err = m.session.GuildMemberRoleAdd(memberLevelStatus.GuildId, memberLevelStatus.MemberId, userRole.RoleId)
		if err != nil {
			log.Printf("[AybushBot::LevelManager] Error on assigning role to member: %v", err)
		}

		m.memberLevelStatuses[memberLevelStatus.MemberId] = &memberLevelStatus
		m.orderedMemberLevelStatuses = append(m.orderedMemberLevelStatuses, &memberLevelStatus)
	}

	sort.Sort(SortedMemberLevelStatuses(m.orderedMemberLevelStatuses))

	go m.workAsync()
}

func (m *Manager) Stop() {
	m.running = false
}

func (m *Manager) workAsync() {
	lastVoicePointsGiven := time.Unix(0, 0)

	for m.running {
		select {
		case user := <- m.onRankQueryChan:
			err := m.rankQueried(user)
			if err != nil {
				log.Printf("[AybushBot::LevelManager] Error on querying the rank of the user: %v", err)
			}
		case memberVoiceChange := <- m.onVoiceChan:
			if memberVoiceChange.shouldEarn {
				m.membersInVoice[memberVoiceChange.memberId] = memberVoiceChange.memberId
			} else {
				delete(m.membersInVoice, memberVoiceChange.memberId)
			}
		case messageCreate := <- m.onMessageChan:
			m.earnExperienceFromMessage(messageCreate)
		default:
			if time.Now().Unix() - lastVoicePointsGiven.Unix() > voiceChannelEarningTimeoutSeconds {
				for _, memberId := range m.membersInVoice {
					m.earnExperienceFromVoice(memberId)
				}

				lastVoicePointsGiven = time.Now()
			}
		}
	}
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
			return
		}
	}

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

	m.earnExperience(status, ExpTypeVoice)
}

func (m *Manager) earnExperience(status *MemberLevelStatus, expType ExpType) {
	earnedExperience := 0

	if expType == ExpTypeVoice {
		earnedExperience = m.calculateEarnedExperience(status.Member, bothSubAndBoosterVoiceMax, bothSubAndBoosterVoiceMin,
			notBoosterButSubVoiceMax, notBoosterButSubVoiceMin, notSubButBoosterVoiceMax, notSubButBoosterVoiceMin, notSubNotBoosterVoiceMax, notSubNotBoosterVoiceMin)
	} else if expType == ExpTypeText {
		earnedExperience = m.calculateEarnedExperience(status.Member,
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

	m.orderedMemberLevelStatusMtx.Lock()
	sort.Sort(SortedMemberLevelStatuses(m.orderedMemberLevelStatuses))
	m.orderedMemberLevelStatusMtx.Unlock()

	log.Printf("[AybushBot::LevelManager] ExperienceType: %v, MemberId: %v, Username: %v#%v, EarnedExp: %v, CurrentLevel: %v, Exp: %v, NextLevel: %v, RequiredExp: %v", expType.String(),
		status.MemberId, status.Member.User.Username, status.Member.User.Discriminator, earnedExperience, status.CurrentLevel.Id, status.ExperiencePoints, status.NextLevel.Id, status.NextLevel.RequiredExperiencePoints)
}

func (m *Manager) calculateEarnedExperience(member *discordgo.Member, bothSubAndBoosterMax int, bothSubAndBoosterMin int, notBoosterButSubMax int, notBoosterButSubMin int,
	notSubButBoosterMax int, notSubButBoosterMin int, notSubNotBoosterMax int, notSubNotBoosterMin int) int {
	earnedExperiencePoints := 0

	isSub, isBooster := false, false

	for _, roleId := range member.Roles {
		if roleId == configuration.Manager.Roles.SubRole {
			isSub = true
		} else if roleId == configuration.Manager.Roles.ServerBoosterRole {
			isBooster = true
		}
	}

	if isSub {
		if isBooster {
			earnedExperiencePoints = rnd.Intn(bothSubAndBoosterMax - bothSubAndBoosterMin) + bothSubAndBoosterMin
		} else {
			earnedExperiencePoints = rnd.Intn(notBoosterButSubMax - notBoosterButSubMin) + notBoosterButSubMin
		}
	} else {
		if isBooster {
			earnedExperiencePoints = rnd.Intn(notSubButBoosterMax - notSubButBoosterMin) + notSubButBoosterMin
		} else {
			earnedExperiencePoints = rnd.Intn(notSubNotBoosterMax - notSubNotBoosterMin) + notSubNotBoosterMin
		}
	}

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
	sort.Sort(SortedMemberLevelStatuses(m.orderedMemberLevelStatuses))
	m.orderedMemberLevelStatusMtx.Unlock()


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

	m.orderedMemberLevelStatusMtx.RLock()
	var currentRank int
	for i, val := range m.orderedMemberLevelStatuses {
		if val.Id == status.Id {
			currentRank = len(m.orderedMemberLevelStatuses) - i
			break
		}
	}
	m.orderedMemberLevelStatusMtx.RUnlock()

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
			Text:          fmt.Sprintf("#%v", currentRank),
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