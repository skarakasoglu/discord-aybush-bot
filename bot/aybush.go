package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	embed "github.com/clinet/discordgo-embed"
	"github.com/skarakasoglu/discord-aybush-bot/bot/antispam"
	"github.com/skarakasoglu/discord-aybush-bot/bot/commands"
	"github.com/skarakasoglu/discord-aybush-bot/bot/level"
	"github.com/skarakasoglu/discord-aybush-bot/configuration"
	"github.com/skarakasoglu/discord-aybush-bot/repository"
	"github.com/skarakasoglu/discord-aybush-bot/shopier/models"
	"github.com/skarakasoglu/discord-aybush-bot/streamlabs"
	slmodels "github.com/skarakasoglu/discord-aybush-bot/streamlabs/models"
	"github.com/skarakasoglu/discord-aybush-bot/twitch"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/messages"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/payloads/v1"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	COMMAND_PREFIX = "!"
	HELP_ARG = "help"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	rnd = rand.New(randSource)
)

type Aybush struct{
	discordConnection *discordgo.Session
	running bool

	levelManager *level.Manager

	antiSpam antispam.AntiSpam
	commands map[string]commands.Command

	userLiveStatuses map[string]bool

	userFollowsChan <-chan v1.UserFollows
	streamChangedChan <-chan messages.StreamChanged
	shopierOrderChan <-chan models.Order

	discordRepository repository.DiscordRepository
	streamlabsApiClient streamlabs.ApiClient

	dmChannel *discordgo.Channel
}

func New(discordConnection *discordgo.Session,
	userFollowChan <-chan v1.UserFollows, streamChangedChan <-chan messages.StreamChanged,
	shopierOrderChan <-chan models.Order, discordRepository repository.DiscordRepository, streamlabsApiClient streamlabs.ApiClient) *Aybush {
	aybus := &Aybush{
		levelManager: level.NewManager(discordConnection, discordRepository, configuration.Manager.LevelSystem.IgnoredTextChannels, configuration.Manager.LevelSystem.IgnoredVoiceChannels),
		discordConnection: discordConnection,
		userFollowsChan: userFollowChan,
		streamChangedChan: streamChangedChan,
		shopierOrderChan: shopierOrderChan,
		discordRepository: discordRepository,
		streamlabsApiClient: streamlabsApiClient,
		userLiveStatuses: make(map[string]bool),
	}

	antiSpamConfiguration := configuration.Manager.AntiSpam
	aybus.antiSpam = antispam.NewAntiSpam(antiSpamConfiguration.MaxInterval, antiSpamConfiguration.MaxDuplicatesInterval,
		configuration.Manager.Roles.ModerationRoles, []string{configuration.Manager.BotUserId}, configuration.Manager.AntiSpam.IgnoredChannels)
	aybus.antiSpam.AddProtectionConfig(antispam.ProtectionConfig{
		Threshold:     antiSpamConfiguration.Mute.Threshold,
		MaxDuplicates: antiSpamConfiguration.Mute.MaxDuplicates,
		Callback:      aybus.muteUserOnSpam,
	})

	aybus.commands = make(map[string]commands.Command)

	joiningDateCmd := commands.NewJoiningDateCommand(discordConnection)
	aybus.commands[joiningDateCmd.Name()] = joiningDateCmd

	clearMsgCmd := commands.NewClearMessageCommand(discordConnection)
	aybus.commands[clearMsgCmd.Name()] = clearMsgCmd

	muteCmd := commands.NewMuteCommand(discordConnection)
	aybus.commands[muteCmd.Name()] = muteCmd

	loveMtrCmd := commands.NewLoveMeterCommand(discordConnection)
	aybus.commands[loveMtrCmd.Name()] = loveMtrCmd

	rockPaperScissors := commands.NewRockPaperScissorsCommand(discordConnection)
	aybus.commands[rockPaperScissors.Name()] = rockPaperScissors

	rank := commands.NewRankCommand(aybus.levelManager.OnRankQuery)
	aybus.commands[rank.Name()] = rank

	leaderboard := commands.NewLeaderboardCommand()
	aybus.commands[leaderboard.Name()] = leaderboard

	//Creating DM channel to aybuse to be able to send shopier order details to her.
	var err error
	aybus.dmChannel, err = aybus.discordConnection.UserChannelCreate("364255114804068352")
	if err != nil {
		log.Printf("[AybushBot] Error on creating DM channel: %v", err)
	}

	return aybus
}

func (a*Aybush) Start() {
	a.running = true

	log.Println("[AybushBot] Registering handlers.")
	a.discordConnection.AddHandler(a.onMemberJoin)
	a.discordConnection.AddHandler(a.onMemberLeave)
	a.discordConnection.AddHandler(a.onMemberUpdate)
	a.discordConnection.AddHandler(a.saveToDatabase)
	a.discordConnection.AddHandler(a.onCommandReceived)
	a.discordConnection.AddHandler(a.onURLSend)
	a.discordConnection.AddHandler(a.onTicketReactionAdd)
	a.discordConnection.AddHandler(a.onTicketReactionRemove)
	a.discordConnection.AddHandler(a.onSpamCheck)
	a.discordConnection.AddHandler(a.onLevel)
	a.discordConnection.AddHandler(a.onVoiceLevel)
	a.discordConnection.AddHandler(a.onRoleCreate)
	a.discordConnection.AddHandler(a.onRoleUpdate)
	a.discordConnection.AddHandler(a.onRoleDelete)
	a.discordConnection.AddHandler(a.onChannelCreate)
	a.discordConnection.AddHandler(a.onChannelUpdate)
	a.discordConnection.AddHandler(a.onChannelDelete)

	a.antiSpam.Start()

	go a.workAsync()
	go a.updatePresence()
	go a.autoBroadcastLeaderboardCommand()
	go a.levelManager.Start()
}

func (a*Aybush) Stop() {
	a.running = false

	a.antiSpam.Stop()
	a.levelManager.Stop()

	err := a.discordConnection.Close()
	if err != nil {
		log.Printf("[AybushBot] Error on closing websocket connection with Discord API: %v", err)
	}
}

func (a *Aybush) workAsync() {
	for a.IsRunning() {
		select {
		case streamChange := <-a.streamChangedChan:
			a.onStreamChanged(streamChange)
		case userFollows := <-a.userFollowsChan:
			a.onUserFollows(userFollows)
		case order := <-a.shopierOrderChan:
			a.onShopierOrderNotify(order)
		}
	}
}

func (a *Aybush) autoBroadcastLeaderboardCommand() {
	leaderboardCommand := commands.NewLeaderboardCommand()

	for a.IsRunning() {
		time.Sleep(time.Duration(2) * time.Hour)

		_, err := a.discordConnection.ChannelMessageSend(configuration.Manager.Channels.Aybus, leaderboardCommand.ResponseMessage())
		if err != nil {
			log.Printf("Error on auto broadcasting leaderboard command: %v", err)
		}
	}
}

func (a *Aybush) IsRunning() bool {
	return a.running
}

func (a *Aybush) updatePresence() {
	for a.IsRunning() {
		for _, val := range configuration.Manager.PresenceUpdate.Statuses {
			err := a.discordConnection.UpdateGameStatus(0, val)
			if err != nil {
				log.Printf("[AybushBot] Error on updating status: %v", err)
			}

			time.Sleep(time.Millisecond * time.Duration(configuration.Manager.PresenceUpdate.PresenceUpdateFrequency))
		}
	}
}

func (a *Aybush) onStreamChanged(streamChange messages.StreamChanged) {
	log.Printf("[AybushBot] Stream changed event received: %v", streamChange)

	if streamChange.Version != twitch.CURRENT_TWITCH_API_VER {
		log.Printf("[AybushBot] Invalid twitch api version. The API version should be %v but got %v.",
			twitch.CURRENT_TWITCH_API_VER, streamChange.Version)
		return
	}

	if streamChange.StreamChangeType == messages.StreamChangeType_Offline {
		log.Printf("[AybushBot] %v ended the stream.", streamChange.Username)
		a.userLiveStatuses[streamChange.Username] = false
		return
	}

	isLive, ok := a.userLiveStatuses[streamChange.Username]
	if ok && isLive {
		log.Printf("[AybushBot] %v is already live, skipping the notification.", streamChange.Username)
		return
	}

	a.userLiveStatuses[streamChange.Username] = true
	twitchUrl := fmt.Sprintf("https://twitch.tv/%v", streamChange.Username)

	embedMsg := embed.NewGenericEmbed(streamChange.Title, "")
	embedMsg.URL = twitchUrl

	thumbnail := strings.Replace(
		strings.Replace(streamChange.ThumbnailURL, "{width}", "400", 1),
		"{height}", "225", 1)

	embedMsg.Author = &discordgo.MessageEmbedAuthor{Name: streamChange.Username, IconURL: streamChange.AvatarURL}
	embedMsg.Thumbnail = &discordgo.MessageEmbedThumbnail{
		URL: streamChange.AvatarURL,
	}
	embedMsg.Color = int(0x6441A4)

	if streamChange.GameName == "" {
		streamChange.GameName = "Just Chatting"
	}

	gameField := &discordgo.MessageEmbedField{
		Name:   "Oyun",
		Value:  streamChange.GameName,
		Inline: true,
	}
	/*
	viewerField := &discordgo.MessageEmbedField{
		Name:   "İzleyiciler",
		Value:  fmt.Sprintf("%v", streamChange.ViewerCount),
		Inline: true,
	}

	 */

	embedMsg.Fields = []*discordgo.MessageEmbedField{gameField}
	embedMsg.Image = &discordgo.MessageEmbedImage{
		URL:      thumbnail,
	}

	_, err := a.discordConnection.ChannelMessageSendComplex(configuration.Manager.Channels.Sohbet, &discordgo.MessageSend{
		Embed: embedMsg,
		Content: fmt.Sprintf("@everyone, %v yayında!.", twitchUrl),
	})
	if err != nil {
		log.Printf("[AybushBot] Error on sending embed message to chat channel: %v", err)
	}
}

func (a *Aybush) onUserFollows(userFollows v1.UserFollows) {
	log.Printf("[AybushBot] User follows event received: %v", userFollows)
	_, err := a.discordConnection.ChannelMessageSend(configuration.Manager.Channels.BotLog,
		fmt.Sprintf("> **%v**: yeni takipçi kazandı. **%v** kullanıcısı **%v** tarihinde takip etti.", userFollows.ToName, userFollows.FromName,
			userFollows.FollowedAt.Local().Format(time.Stamp)))
	if err != nil {
		log.Printf("[AybushBot] Error on writing to bot log channel: %v", err)
	}
}

func (a *Aybush) onShopierOrderNotify(order models.Order) {
	price, err := strconv.ParseFloat(order.Price, 64)
	if err != nil {
		log.Printf("[AybushBot] Error on parsing price to float: %v", err)
	}

	donationId, err := a.streamlabsApiClient.CreateDonation(slmodels.CreateDonation{
		Name:       fmt.Sprintf("%v %v", order.Name, order.Surname),
		Message:    fmt.Sprintf("%v IDli üründen %v adet satın aldı.", order.ProductId, order.ProductCount),
		Identifier: order.Email,
		Amount:     price,
		CreatedAt:  time.Now().Unix(),
		Currency:   order.CurrencyString,
		SkipAlert:  "yes",
	})
	if err != nil {
		log.Printf("[AybushBot] Error on creating donation in streamlabs: %v", err)
	}
	log.Printf("[AybushBot] Donation was created successfully. DonationId: %v", donationId)

	_, err = a.streamlabsApiClient.CreateAlert(slmodels.CreateAlert{
		Type:             slmodels.AlertType_Donation,
		ImageHref:        fmt.Sprint("https://shopier.aybushbot.com/images/venom.png"),
		SoundHref:        fmt.Sprint("https://shopier.aybushbot.com/alerts/order_alert.mp3"),
		Message:          fmt.Sprint("Bir ürün satın alındı!"),
		UserMessage:      "Çok teşekkürler aybuseMutlu",
		Duration:         11000,
		SpecialTextColor: "Pink",
	})
	if err != nil {
		log.Printf("[AybushBot] Error on creating alert in streamlabs: %v", err)
	}

	if a.dmChannel != nil {
		messageStr := fmt.Sprintf(`**Sipariş Kimliği:** %v
**Sipariş Veren:** %v %v
**Siparişi Veren E-Mail:** %v
**Ürün Kimliği:** %v
**Adet:** %v
**Fiyat:** %v
**Müşteri Notu:** %v`, order.OrderId, order.Name, order.Surname, order.Email, order.ProductId, order.ProductCount, order.Price, order.CustomerNote)

		_, err = a.discordConnection.ChannelMessageSend(a.dmChannel.ID, messageStr)
		if err != nil {
			log.Printf("[AybushBot] Error on sending message to channel: %v", err)
		}
	}
}