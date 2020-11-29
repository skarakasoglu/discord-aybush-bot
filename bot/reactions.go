package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/skarakasoglu/discord-aybush-bot/configuration"
	"log"
)

func (a *Aybus) onTicketReactionAdd(session *discordgo.Session, reactionAdd *discordgo.MessageReactionAdd) {
	if reactionAdd.MessageID != configuration.Manager.Ticket.MessageId || reactionAdd.Emoji.Name != configuration.Manager.Ticket.Reaction {
		return
	}

	guild, err := session.Guild(reactionAdd.GuildID)
	if err != nil {
		log.Printf("Error on obtaining guild: %v", err)
	}

	member, err := session.GuildMember(reactionAdd.GuildID, reactionAdd.UserID)
	if err != nil {
		log.Printf("Error on obtaining guild member: %v", err)
	}

	log.Printf("Assigning ticket role to %v#%v in %v.", member.User.Username, member.User.Discriminator, guild.Name)

	err = session.GuildMemberRoleAdd(guild.ID, member.User.ID, configuration.Manager.Ticket.RoleId)
	if err != nil {
		log.Printf("Error on adding ticket role to member: %v", err)
	}
}

func (a *Aybus) onTicketReactionRemove(session *discordgo.Session, reactionRemove *discordgo.MessageReactionRemove) {
	if reactionRemove.MessageID != configuration.Manager.Ticket.MessageId || reactionRemove.Emoji.Name != configuration.Manager.Ticket.Reaction {
		return
	}

	guild, err := session.Guild(reactionRemove.GuildID)
	if err != nil {
		log.Printf("Error on obtaining guild: %v", err)
	}

	member, err := session.GuildMember(reactionRemove.GuildID, reactionRemove.UserID)
	if err != nil {
		log.Printf("Error on obtaining guild member: %v", err)
	}

	log.Printf("Removing ticket role from %v#%v in %v.", member.User.Username, member.User.Discriminator, guild.Name)

	err = session.GuildMemberRoleRemove(guild.ID, member.User.ID, configuration.Manager.Ticket.RoleId)
	if err != nil {
		log.Printf("Error on removing ticket role to member: %v", err)
	}
}