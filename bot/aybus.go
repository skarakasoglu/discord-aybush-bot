package bot

import (
	"github.com/bwmarrin/discordgo"
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
	welcomeMessages []string
}

func New(discordConnection *discordgo.Session) *Aybus{
	welcomeMessages := []string{
		"Hoşgeldin",
		"Sa",
	}

	return &Aybus{
		discordConnection: discordConnection,
		welcomeMessages: welcomeMessages,
	}
}

func (a* Aybus) Start() {
	log.Println("Registering handlers.")
	a.discordConnection.AddHandler(a.handleWelcomeMessage)
	//a.discordConnection.AddHandler(a.handleGeneralMessage)

	log.Println("Aybus started.")
}

func (a* Aybus) Stop() {

}

// Handlers
func (a* Aybus) handleGeneralMessage(session *discordgo.Session, message *discordgo.MessageCreate) {
	log.Println(message.Message.Content)
}

func (a* Aybus) handleWelcomeMessage(session *discordgo.Session, memberAdd *discordgo.GuildMemberAdd) {
	log.Println("Test")
	msgIndex := rnd.Intn(len(a.welcomeMessages))

	log.Printf("%v joined. WelcomeMessage: %v", memberAdd.User.Username, a.welcomeMessages[msgIndex])

}

