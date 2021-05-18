package twitch

import (
	"fmt"
	twitchirc "github.com/gempir/go-twitch-irc/v2"
	"github.com/skarakasoglu/discord-aybush-bot/data/models"
	"github.com/skarakasoglu/discord-aybush-bot/repository"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/payloads"
	"log"
	"math/rand"
	"strings"
	"time"
)

const (
	chatInactiveSeconds = 300
)

var (
	randSrc = rand.NewSource(time.Now().UnixNano())
	rnd = rand.New(randSrc)
)

type ChatBot struct {
	running bool

	inactiveMode bool

	username string
	token string
	client *ApiClient

	ircClient *twitchirc.Client
	twitchRepository repository.TwitchRepository

	chatCommands []models.TwitchBotCommand
	noticeEventMessages map[string]map[string]map[bool]models.TwitchBotUserNoticeMessage
	autoBroadcastMessages []models.TwitchBotAutoBroadcastMessage
	bitMessages []models.TwitchBotMessage

	streamer payloads.User

	chatActive chan bool
	lastMessageTimestamp time.Time
}

func NewChatBot(username string, token string, streamer payloads.User, client *ApiClient, twitchRepository repository.TwitchRepository) *ChatBot {
	return &ChatBot{
		running: false,
		inactiveMode: false,
		username: username,
		token:    fmt.Sprintf("oauth:%v", token),
		streamer: streamer,
		client: client,
		ircClient: twitchirc.NewClient(username, fmt.Sprintf("oauth:%v", token)),
		twitchRepository: twitchRepository,
		noticeEventMessages: make(map[string]map[string]map[bool]models.TwitchBotUserNoticeMessage),
		chatActive: make(chan bool, 5),
	}
}

func (cb *ChatBot) Start() {
	cb.running = true

	var err error
	cb.chatCommands, err = cb.twitchRepository.GetAllTwitchBotCommands()
	if err != nil {
		log.Printf("[TwitchChatBot] Error on receiving twitch bot commands: %v", err)
	}

	userNoticeMessages, err := cb.twitchRepository.GetAllTwitchBotUserNoticeMessages()
	if err != nil {
		log.Printf("[TwitchChatBot] Error on receiving twitch bot user notice messages: %v", err)
	}

	cb.autoBroadcastMessages, err = cb.twitchRepository.GetAllTwitchBotAutoBroadcastMessages()
	if err != nil {
		log.Printf("[TwitchChatBot] Error on obtaining auto broadcast messages: %v", err)
	}

	for _, noticeMessage := range userNoticeMessages {
		noticeEvent, ok := cb.noticeEventMessages[noticeMessage.NoticeEvent]
		if !ok {
			cb.noticeEventMessages[noticeMessage.NoticeEvent] = make(map[string]map[bool]models.TwitchBotUserNoticeMessage)
			noticeEvent = cb.noticeEventMessages[noticeMessage.NoticeEvent]
		}

		tier, ok := noticeEvent[noticeMessage.Tier]
		if !ok {
			noticeEvent[noticeMessage.Tier] = make(map[bool]models.TwitchBotUserNoticeMessage)
			tier = noticeEvent[noticeMessage.Tier]
		}

		tier[noticeMessage.IsRecipientMe] = noticeMessage
	}

	botMessages, err := cb.twitchRepository.GetAllTwitchBotMessages()
	if err != nil {
		log.Printf("[TwitchChatBot] Error on obtaining twitch bot messages: %v", err)
	}

	for _, message := range botMessages {
		if message.MinimumBits > 0 {
			cb.bitMessages = append(cb.bitMessages, message)
		}
	}

	go func() {
		for cb.running {
			if time.Now().Unix() - cb.lastMessageTimestamp.Unix() >= chatInactiveSeconds {
				log.Printf("[TwitchChatBot] Bot is in inactive mode right now.")
				cb.inactiveMode = true

				for range cb.chatActive {
					cb.inactiveMode = false
					break
				}
				log.Printf("[TwitchChatBot] Bot is getting in active mode again.")
			}

			for _, message := range cb.autoBroadcastMessages {
				log.Printf("[TwitchChatBot] AutoBroadcastMessage: Content: %v, IntervalSeconds: %v", message.Message.Content, message.IntervalSeconds)
				cb.ircClient.Say(cb.streamer.Login, message.Message.Content)
				time.Sleep(time.Second * time.Duration(message.IntervalSeconds))
			}
		}
	}()

	cb.ircClient.OnUserNoticeMessage(cb.onUserNotice)
	cb.ircClient.OnPrivateMessage(func(message twitchirc.PrivateMessage) {
		log.Printf("[TwitchChatBot] sender=%v, sent a message to the channel %v: %v", message.User, message.Channel, message.Message)
		cb.lastMessageTimestamp = time.Now()

		if cb.inactiveMode {
			cb.chatActive <- true
		}

		cb.onBitsReceived(&message)
		cb.onFollowage(&message)
		cb.onCommandReceived(&message)
	})

	cb.ircClient.Join(cb.streamer.Login)

	go func() {
		err = cb.ircClient.Connect()
		if err != nil {
			panic(err)
		}
	}()

}

func (cb *ChatBot) Stop() {
	cb.running = false
}

func (cb *ChatBot) onUserNotice(message twitchirc.UserNoticeMessage) {
	msg := ""
	tier := message.MsgParams["msg-param-sub-plan"]

	switch message.MsgID {
	case "resub":
		cumulativeMonths := message.MsgParams["msg-param-cumulative-months"]

		noticeEventMessage := cb.noticeEventMessages[message.MsgID][tier][false]
		msg = fmt.Sprintf(noticeEventMessage.Content, message.User.DisplayName, cumulativeMonths)
	case "sub":
		noticeEventMessage := cb.noticeEventMessages[message.MsgID][tier][false]

		msg = fmt.Sprintf(noticeEventMessage.Content, message.User.DisplayName)
	case "subgift":
		gifted := message.MsgParams["msg-param-recipient-user-name"]
		streak := message.MsgParams["msg-param-months"]

		noticeEventMessage := cb.noticeEventMessages[message.MsgID][tier][gifted == cb.username]

		msg = fmt.Sprintf(noticeEventMessage.Content, message.User.DisplayName, streak)
	}

	if msg != "" {
		cb.ircClient.Say(message.Channel, msg)
	}
}

func (cb *ChatBot) onBitsReceived(message *twitchirc.PrivateMessage) {
	if message.Bits > 0 {
		cb.ircClient.Say(message.Channel, fmt.Sprintf(cb.bitMessages[0].Content, message.User.DisplayName, message.Bits))
	}
}

func (cb *ChatBot) onCommandReceived(message *twitchirc.PrivateMessage) {
	var anyCommand *models.TwitchBotCommand
	for _, val := range cb.chatCommands {
		if strings.HasPrefix(strings.ToLower(message.Message), fmt.Sprintf("%v ", strings.ToLower(val.Command))) || strings.ToLower(val.Command) == strings.ToLower(message.Message) {
			anyCommand = &val
			log.Printf("[TwitchChatBot] sender=%+v, message: %+v, contains a command: %+v", message.User, message.Message, val)
			break
		}
	}

	if anyCommand != nil {
		if anyCommand.Message.Type.Name == "Hoşgeldin" {
			username := strings.ToLower(message.User.Name)

			if username == "crossman" || username == "meldabaker" || username == "bidik" ||
				username == "helengun_" || username == "ceydoouu" || username == "segfaultc" || username == "aybusee" ||
				username == "mithzim" {
				cb.ircClient.Say(message.Channel, fmt.Sprintf("@%v, aleyküm selam lan berkay...", message.User.DisplayName))
				return
			}
		}

		cb.ircClient.Say(message.Channel, fmt.Sprintf("@%v, %v", message.User.DisplayName, anyCommand.Message.Content))
	}
}

func (cb *ChatBot) onFollowage(message *twitchirc.PrivateMessage) {
	if strings.HasPrefix(message.Message, "!followage ") || message.Message == "!followage" {
		followageData := cb.client.getUserFollowage(message.User.ID, cb.streamer.Id)

		if followageData.FromID == "" {
			cb.ircClient.Say(message.Channel, fmt.Sprintf("@%v beni takip etmiyor. aybuseAgla", message.User.DisplayName))
			return
		}

		followageTime := time.Now().Sub(followageData.FollowedAt)

		totalFollowageDays := int(followageTime.Hours() / 24)
		followageYears := totalFollowageDays / 365
		followageMonths := (totalFollowageDays % 365) / 30
		followageDays := (totalFollowageDays % 365) % 30

		followageHours := int(followageTime.Hours()) % 24
		followageMinutes := int(followageTime.Minutes()) % 60

		builder := strings.Builder{}

		builder.WriteString(fmt.Sprintf("@%v", message.User.DisplayName))

		if followageYears > 0 {
			builder.WriteString(fmt.Sprintf(" %v yıl", followageYears))
		}

		if followageMonths > 0 {
			builder.WriteString(fmt.Sprintf(" %v ay", followageMonths))
		}

		if followageDays > 0 {
			builder.WriteString(fmt.Sprintf(" %v gün", followageDays))
		}

		if followageHours > 0 {
			builder.WriteString(fmt.Sprintf(" %v saat", followageHours))
		}

		if followageMinutes > 0 {
			builder.WriteString(fmt.Sprintf(" %v dakika", followageMinutes))
		}

		if strings.HasSuffix(builder.String(), "yıl") || strings.HasSuffix(builder.String(), "ay") || strings.HasSuffix(builder.String(), "dakika") {
			builder.WriteString("dır ")
		} else if strings.HasSuffix(builder.String(), "gün") {
			builder.WriteString("dür ")
		}else if strings.HasSuffix(builder.String(), "saat") {
			builder.WriteString("tir ")
		}
		builder.WriteString("beni takip ediyor. Teşekkürler minik Ragga Muffin aybuseMutlu")

		cb.ircClient.Say(message.Channel, builder.String())
	}
}