package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/skarakasoglu/discord-aybush-bot/configuration"
	"log"
)

func (a* Aybus) onMemberJoin(session *discordgo.Session, memberAdd *discordgo.GuildMemberAdd) {
	msgIndex := rnd.Intn(len(configuration.Manager.Greeting.GreetingMessages))

	guild, err := session.Guild(memberAdd.GuildID)
	if err != nil {
		log.Printf("Error on obtaining guild: %v", err)
	}

	log.Printf("%v joined to %v.", memberAdd.User.Username, guild.Name)

	msg := fmt.Sprintf("<@%v>, **Hoşgeldin.** <:aybuscorona:692160124143403008> %v", memberAdd.User.ID, configuration.Manager.Greeting.GreetingMessages[msgIndex])

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

	log.Printf("%v left from %v", memberLeave.User.Username, guild.Name)

	botLogMsg := fmt.Sprintf("> **ID**: `%v`, **Kullanıcı Adı**: `%v#%v` sunucudan ayrıldı.",
		memberLeave.User.ID, memberLeave.User.Username, memberLeave.User.Discriminator)

	_, err = session.ChannelMessageSend(configuration.Manager.Channels.BotLog, botLogMsg)
	if err != nil {
		log.Printf("Error on logging to bot log channel: %v", err)
	}
}