package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/skarakasoglu/discord-aybush-bot/data/models"
	"log"
	"time"
)

func (a *Aybush) onChannelCreate(session *discordgo.Session, create *discordgo.ChannelCreate) {
	if create.Type == discordgo.ChannelTypeGuildText {
		textChannel := models.DiscordTextChannel{
			ChannelId: create.Channel.ID,
			Name:      create.Channel.Name,
			IsNsfw:    create.Channel.NSFW,
			CreatedAt: time.Now(),
		}

		log.Printf("New text channel created. ChannelId: %v, Name: %v", textChannel.ChannelId, textChannel.Name)

		_, err := a.discordRepository.InsertDiscordTextChannel(textChannel)
		if err != nil {
			log.Printf("Error on inserting discord text channel: %v", err)
		}
	}
}

func (a *Aybush) onChannelUpdate(session *discordgo.Session, update *discordgo.ChannelUpdate) {
	if update.Type == discordgo.ChannelTypeGuildText {
		textChannel := models.DiscordTextChannel{
			ChannelId: update.Channel.ID,
			Name:      update.Channel.Name,
			IsNsfw:    update.Channel.NSFW,
		}

		log.Printf("Text channel updated. ChannelId: %v, Name: %v", textChannel.ChannelId, textChannel.Name)

		_, err := a.discordRepository.UpdateDiscordTextChannelById(textChannel)
		if err != nil {
			log.Printf("Error on inserting discord text channel: %v", err)
		}
	}
}

func (a *Aybush) onChannelDelete(session *discordgo.Session, channelDelete *discordgo.ChannelDelete) {
	if channelDelete.Type == discordgo.ChannelTypeGuildText {
		log.Printf("Text channel deleted. ChannelId: %v, Name: %v", channelDelete.ID, channelDelete.Name)

		_, err := a.discordRepository.DeleteDiscordTextChannelById(channelDelete.ID)
		if err != nil {
			log.Printf("Error on inserting discord text channel: %v", err)
		}
	}
}