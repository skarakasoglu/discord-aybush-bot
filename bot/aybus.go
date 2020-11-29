package bot

import (
	"fmt"
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
	a.discordConnection.AddHandler(a.onTicketReactionAdd)
	a.discordConnection.AddHandler(a.onTicketReactionRemove)

	go a.updatePresence()
	log.Println("Aybus has been started.")
}

func (a* Aybus) Stop() {
	a.running = false
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

// Handlers
func (a* Aybus) onMemberJoin(session *discordgo.Session, memberAdd *discordgo.GuildMemberAdd) {
	msgIndex := rnd.Intn(len(configuration.Manager.Greeting.GreetingMessages))

	guild, err := session.Guild(memberAdd.GuildID)
	if err != nil {
		log.Printf("Error on obtaining guild: %v", err)
	}

	log.Printf("%v joined to %v.", memberAdd.User.Username, guild.Name)

	msg := fmt.Sprintf("<@%v> %v", memberAdd.User.ID, configuration.Manager.Greeting.GreetingMessages[msgIndex])

	_, err = session.ChannelMessageSend(configuration.Manager.Greeting.GreetingChannel, msg)
	if err != nil {
		log.Printf("Error on sending message to greetings channel: %v", err)
	}

	err = session.GuildMemberRoleAdd(memberAdd.GuildID, memberAdd.User.ID, configuration.Manager.Roles.DefaultMemberRole)
	if err != nil {
		log.Printf("Error on assigning default role: %v", err)
	}

	dmChannel, err := session.UserChannelCreate(memberAdd.User.ID)
	if err != nil {
		log.Printf("Error on creating DM channel: %v", err)
	}

	_, err = session.ChannelMessageSend(dmChannel.ID, configuration.Manager.Greeting.GreetingDirectMessage)
	if err != nil {
		log.Printf("Error on sending message via DM channel: %v", err)
	}

}

func (a* Aybus) onMemberLeave(session *discordgo.Session, memberLeave *discordgo.GuildMemberRemove) {
	guild, err := session.Guild(memberLeave.GuildID)
	if err != nil {
		log.Printf("Error on obtaining guild: %v", err)
	}

	fmt.Printf("%v left from %v", memberLeave.User.Username, guild.Name)

	botLogMsg := fmt.Sprintf("> **ID**: %v, **Kullanıcı Adı**: %v#%v sunucudan ayrıldı.",
		memberLeave.User.ID, memberLeave.User.Username, memberLeave.User.Discriminator)

	_, err = session.ChannelMessageSend(configuration.Manager.Channels.BotLog, botLogMsg)
	if err != nil {
		log.Printf("Error on logging to bot log channel: %v", err)
	}
}
