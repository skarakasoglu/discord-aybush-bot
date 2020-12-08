package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	embed "github.com/clinet/discordgo-embed"
	"github.com/skarakasoglu/discord-aybush-bot/configuration"
	"log"
	"math/rand"
	"strings"
	"time"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	rnd = rand.New(randSource)
)

type loveMeterCommand struct{
	session *discordgo.Session
}


func NewLoveMeterCommand(session *discordgo.Session) Command{
	return &loveMeterCommand{
		session: session,
	}
}

func (cmd *loveMeterCommand) Name() string{
	return "aşk-ölçer"
}

func (cmd *loveMeterCommand) Execute(message *discordgo.Message) (string, error){
	if message.ChannelID != configuration.Manager.Channels.Aybus{
		log.Printf("%v command is received in wrong channel. User: %v#%v, channel: %v", cmd.Name(),
			message.Author.Username, message.Author.Discriminator, message.ChannelID) 
		return "", nil
	}

	if len(message.Mentions) != 1{
		return fmt.Sprintf("<@%v>, %v", message.Author.ID, cmd.Usage()), nil
	}

	member,err := cmd.session.GuildMember(message.GuildID, message.Mentions[0].ID)
	if err != nil{
		log.Printf("Error on optaining member: %v", err)
		return "", err
	}

	// aybus.go 'dan "rnd"yi çekemedim sıfırdan burada random oluşturdum. Olmadı sen onu düzenlersin
	love_point := rnd.Intn(100)
	heart := ":heart:"
	bheart := ":black_heart:"
	heart = strings.Repeat(heart, love_point/10)
	bheart = strings.Repeat(bheart, (100 - love_point) / 10 + 1 )

	// Love meter text
	textIndex := int( love_point / 10)
	text := configuration.Manager.LoveMeter.Texts[textIndex]

	if member.User.ID == "621427811869261826" && message.Author.ID == "763734826422370305" || member.User.ID == "763734826422370305" && message.Author.ID == "621427811869261826" {
		text = "Sizi anca zaten birbiriniz bu kadar sevebilir..."
		love_point = int(1000)
	}

	if member.User.ID == "534162038751232000" && message.Author.ID == "482955098415431680" || member.User.ID == "482955098415431680" && message.Author.ID == "534162038751232000" {
		text = "BU AŞK ADETA ÖPÜŞMELİ VURUŞMALI AŞK FİLMLERİNDEN ÇIKMIŞ. İNANILMAZ"
		love_point = int(150)
	}

	// Creating new embed
	newEmbed := embed.NewGenericEmbed("","")
	newEmbed.Title = fmt.Sprintf("%v#%v ile %v#%v Arasındaki Aşk Sonucu", message.Author.Username,message.Author.Discriminator,member.User.Username,member.User.Discriminator)
	newEmbed.Color = int(0xBE1931)
	heartsField := discordgo.MessageEmbedField{Name: fmt.Sprintf("**Aşk Yüzdesi:** %v", love_point), Value: fmt.Sprintf("%v%v\n\n%v", heart, bheart, text), Inline: false}
	newEmbed.Fields = []*discordgo.MessageEmbedField{&heartsField}

	_, err = cmd.session.ChannelMessageSendEmbed(message.ChannelID, newEmbed)
	if err != nil {
		log.Printf("Error sending embed message Aybus Channel: %v", err)
	}

	log.Printf("%v between %v",member.User.Username,message.Author.Username)
	return  "", nil
}

func (cmd *loveMeterCommand) Usage() string{
	return "**bu komutu,**\n> !aşk-ölçer `<@kullanıcı-adı>`\nşeklinde kullanabilirsiniz."
}
