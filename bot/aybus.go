package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/skarakasoglu/discord-aybush-bot/bot/antispam"
	"github.com/skarakasoglu/discord-aybush-bot/bot/commands"
	"github.com/skarakasoglu/discord-aybush-bot/configuration"
	"log"
	"math/rand"
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

type Aybus struct{
	discordConnection *discordgo.Session
	running bool

	antiSpam antispam.AntiSpam
	commands map[string]commands.Command
}

func New(discordConnection *discordgo.Session) *Aybus{
	aybus := &Aybus{
		discordConnection: discordConnection,
	}

	antiSpamConfiguration := configuration.Manager.AntiSpam
	aybus.antiSpam = antispam.NewAntiSpam(antiSpamConfiguration.MaxInterval, antiSpamConfiguration.MaxDuplicatesInterval,
		configuration.Manager.Roles.ModerationRoles, []string{configuration.Manager.BotRoleId})
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

	return aybus
}

func (a* Aybus) Start() {
	a.running = true

	log.Println("Registering handlers.")
	a.discordConnection.AddHandler(a.onMemberJoin)
	a.discordConnection.AddHandler(a.onMemberLeave)
	a.discordConnection.AddHandler(a.onCommandReceived)
	a.discordConnection.AddHandler(a.onURLSend)
	a.discordConnection.AddHandler(a.onTicketReactionAdd)
	a.discordConnection.AddHandler(a.onTicketReactionRemove)
	a.discordConnection.AddHandler(a.onSpamCheck)

	go a.updatePresence()
}

func (a* Aybus) Stop() {
	a.running = false

	err := a.discordConnection.Close()
	if err != nil {
		log.Printf("Error on closing websocket connection with Discord API: %v", err)
	}
}

func (a *Aybus) IsRunning() bool {
	return a.running
}

func (a *Aybus) updatePresence() {
	for a.IsRunning() {
		for _, val := range configuration.Manager.PresenceUpdate.Statuses {
			err := a.discordConnection.UpdateStatus(0, val)
			if err != nil {
				log.Printf("Error on updating status: %v", err)
			}

			time.Sleep(time.Millisecond * time.Duration(configuration.Manager.PresenceUpdate.PresenceUpdateFrequency))
		}
	}
}
