package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/skarakasoglu/discord-aybush-bot/configuration"
	"github.com/skarakasoglu/discord-aybush-bot/data/models"
	"log"
)

func (a*Aybush) onMemberJoin(session *discordgo.Session, memberAdd *discordgo.GuildMemberAdd) {
	msgIndex := rnd.Intn(len(configuration.Manager.Greeting.GreetingMessages))

	guild, err := session.Guild(memberAdd.GuildID)
	if err != nil {
		log.Printf("[AybushBot] Error on obtaining guild: %v", err)
	}

	log.Printf("[AybushBot] MemberId: %v, Username: %v#%v joined to %v.", memberAdd.User.ID, memberAdd.User.Username, memberAdd.User.Discriminator, guild.Name)

	msg := fmt.Sprintf("<@%v>, **Hoşgeldin.** <:aybuscorona:692160124143403008> %v", memberAdd.User.ID, configuration.Manager.Greeting.GreetingMessages[msgIndex])

	_, err = session.ChannelMessageSend(configuration.Manager.Greeting.GreetingChannel, msg)
	if err != nil {
		log.Printf("[AybushBot] Error on sending message to greetings channel: %v", err)
	}

	err = session.GuildMemberRoleAdd(memberAdd.GuildID, memberAdd.User.ID, configuration.Manager.Roles.DefaultMemberRole)
	if err != nil {
		log.Printf("[AybushBot] Error on assigning default role: %v", err)
	}

	dmChannel, err := session.UserChannelCreate(memberAdd.User.ID)
	if err != nil {
		log.Printf("[AybushBot] Error on creating DM channel: %v", err)
	}

	_, err = session.ChannelMessageSend(dmChannel.ID, configuration.Manager.Greeting.GreetingDirectMessage)
	if err != nil {
		log.Printf("[AybushBot] Error on sending message via DM channel: %v", err)
	}

	joinedAt, err := memberAdd.JoinedAt.Parse()
	if err != nil {
		log.Printf("[AybushBot] Error on parsing joined at: %v", err)
	}

	member := models.DiscordMember{
		Id:            0,
		MemberId:      memberAdd.User.ID,
		Email:         memberAdd.User.Email,
		Username:      memberAdd.User.Username,
		Discriminator: memberAdd.User.Discriminator,
		IsVerified:    memberAdd.User.Verified,
		IsBot:         memberAdd.User.Bot,
		Left:          false,
		JoinedAt:      joinedAt,
		GuildId: memberAdd.GuildID,
	}

	_, err = a.discordRepository.InsertDiscordMember(member)
	if err != nil {
		log.Printf("[AybushBot] Error on inserting new member: %v to database: %v", member, err)
	}
}

func (a*Aybush) onMemberLeave(session *discordgo.Session, memberLeave *discordgo.GuildMemberRemove) {
	guild, err := session.Guild(memberLeave.GuildID)
	if err != nil {
		log.Printf("[AybushBot] Error on obtaining guild: %v", err)
	}

	log.Printf("%v left from %v", memberLeave.User.Username, guild.Name)

	botLogMsg := fmt.Sprintf("> **ID**: `%v`, **Kullanıcı Adı**: `%v#%v` sunucudan ayrıldı.",
		memberLeave.User.ID, memberLeave.User.Username, memberLeave.User.Discriminator)

	_, err = session.ChannelMessageSend(configuration.Manager.Channels.BotLog, botLogMsg)
	if err != nil {
		log.Printf("[AybushBot] Error on logging to bot log channel: %v", err)
	}

	member, err := a.discordRepository.GetDiscordMemberById(memberLeave.User.ID)
	if err != nil {
		log.Printf("[AybushBot] Error on obtaining discord member: %v", err)
		return
	}

	member.Left = true
	_, err = a.discordRepository.UpdateDiscordMemberById(member)
	if err != nil {
		log.Printf("[AybushBot] Error on updating discord member: %v", err)
	}
}