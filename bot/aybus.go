package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/skarakasoglu/discord-aybush-bot/configuration"
	"log"
	"math/rand"
	"time"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	rnd = rand.New(randSource)
)

type Aybus struct{
	discordConnection *discordgo.Session
	running bool
}

func New(discordConnection *discordgo.Session) *Aybus{
	return &Aybus{
		discordConnection: discordConnection,
	}
}

func (a* Aybus) Start() {
	a.running = true

	log.Println("Registering handlers.")
	a.discordConnection.AddHandler(a.onMemberJoin)
	a.discordConnection.AddHandler(a.onMemberLeave)
	a.discordConnection.AddHandler(a.onURLSend)
	a.discordConnection.AddHandler(a.onTicketReactionAdd)
	a.discordConnection.AddHandler(a.onTicketReactionRemove)

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
