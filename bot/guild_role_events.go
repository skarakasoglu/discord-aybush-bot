package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/skarakasoglu/discord-aybush-bot/data/models"
	"log"
)

func (a *Aybush) onRoleCreate(session *discordgo.Session, create *discordgo.GuildRoleCreate) {
	role := models.DiscordRole{
		RoleId: create.Role.ID,
		Name:   create.Role.Name,
	}
	_, err := a.discordRepository.InsertDiscordRole(role)
	if err != nil {
		log.Printf("[AybushBot] Error on inserting new discord role: %v", err)
	}
}

func (a *Aybush) onRoleUpdate(session *discordgo.Session, update *discordgo.GuildRoleUpdate) {
	role := models.DiscordRole{
		RoleId: update.Role.ID,
		Name:   update.Role.Name,
	}

	_, err := a.discordRepository.UpdateDiscordRoleById(role)
	if err != nil {
		log.Printf("[AybushBot] Error on updating discord role: %v", err)
	}
}

func (a *Aybush) onRoleDelete(session *discordgo.Session, delete *discordgo.GuildRoleDelete) {
	_, err := a.discordRepository.DeleteDiscordRoleById(delete.RoleID)
	if err != nil {
		log.Printf("[AybushBot] Error on updating discord role: %v", err)
	}
}
