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
	"github.com/skarakasoglu/discord-aybush-bot/twitch/messages"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/payloads"
	"log"
	"math/rand"
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

	userFollowsChan <-chan payloads.UserFollows
	streamChangedChan <-chan messages.StreamChanged

	discordRepository repository.DiscordRepository
}

func New(discordConnection *discordgo.Session,
	userFollowChan <-chan payloads.UserFollows, streamChangedChan <-chan messages.StreamChanged, discordRepository repository.DiscordRepository) *Aybush {
	aybus := &Aybush{
		levelManager: level.NewManager(discordConnection, discordRepository, configuration.Manager.LevelSystem.IgnoredTextChannels, configuration.Manager.LevelSystem.IgnoredVoiceChannels),
		discordConnection: discordConnection,
		userFollowsChan: userFollowChan,
		streamChangedChan: streamChangedChan,
		discordRepository: discordRepository,
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

	return aybus
}

func (a*Aybush) Start() {
	a.running = true

	log.Println("[AybushBot] Registering handlers.")
	a.discordConnection.AddHandler(a.onMemberJoin)
	a.discordConnection.AddHandler(a.onMemberLeave)
	a.discordConnection.AddHandler(a.onMemberUpdate)
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

	go a.levelManager.Start()
	go a.updatePresence()
	go a.receiveStreamChanges()
	go a.receiveUserFollows()
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

func (a *Aybush) receiveStreamChanges() {
	isLive := false

	for a.IsRunning() {
		for streamChange := range a.streamChangedChan {
			log.Printf("[AybushBot] Stream changed event received: %v", streamChange)

			if streamChange.UserID == "0" {
				log.Printf("[AybushBot] %v ended the stream.", streamChange.Username)
				isLive = false
				continue
			}

			if isLive {
				continue
			}

			isLive = true
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

			gameField := &discordgo.MessageEmbedField{
				Name:   "Oyun",
				Value:  streamChange.GameName,
				Inline: true,
			}
			viewerField := &discordgo.MessageEmbedField{
				Name:   "İzleyiciler",
				Value:  fmt.Sprintf("%v", streamChange.ViewerCount),
				Inline: true,
			}

			embedMsg.Fields = []*discordgo.MessageEmbedField{gameField, viewerField}
			embedMsg.Image = &discordgo.MessageEmbedImage{
				URL:      thumbnail,
			}

			_, err := a.discordConnection.ChannelMessageSendComplex(configuration.Manager.Channels.Sohbet, &discordgo.MessageSend{
				Embed: embedMsg,
				Content: fmt.Sprintf("@everyone, %v yayında! Gel gel gel Aybuse'ye gel.", twitchUrl),
			})
			if err != nil {
				log.Printf("[AybushBot] Error on sending embed message to chat channel: %v", err)
			}
		}
	}
}

func (a *Aybush) receiveUserFollows() {
	for a.IsRunning() {
		for userFollows := range a.userFollowsChan {
			log.Printf("[AybushBot] User follows event received: %v", userFollows)
			_, err := a.discordConnection.ChannelMessageSend(configuration.Manager.Channels.BotLog,
				fmt.Sprintf("> **%v** aybusee'yi **%v** tarihinde takip etti.", userFollows.FromName,
					userFollows.FollowedAt.Local().Format(time.Stamp)))
			if err != nil {
				log.Printf("[AybushBot] Error on writing to bot log channel: %v", err)
			}
		}
	}
}