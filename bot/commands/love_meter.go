package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	embed "github.com/clinet/discordgo-embed"
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

	_, err = cmd.session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("%v - %v%v",love_point, heart,bheart))
	if err != nil {
		log.Printf("Error sending message Aybus Channel: %v", err)
	}

	// Creating new embed
	newEmbed := embed.NewGenericEmbed("","")
	newEmbed.Title = fmt.Sprintf("%v#%v ile %v#%v Arasındaki Aşk Sonucu", message.Author.Username,message.Author.Discriminator,member.User.Username,member.User.Discriminator)
	newEmbed.Color = int(0xBE1931)

	// ÖNEMLİ: Fieldlerde sıkıntı oldu anlamadım nedenini. Neyse benim ara  vermem gerekti o yüzden böyle deployladım.
	newEmbedField := discordgo.MessageEmbedField{Name: fmt.Sprintf("**Aşy Yüzdesi:** %v", love_point)}
	newEmbed.Fields = []*discordgo.MessageEmbedField{&newEmbedField}

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
