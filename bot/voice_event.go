package bot

import "github.com/bwmarrin/discordgo"

func (a*Aybush) onVoiceLevel(session *discordgo.Session, update *discordgo.VoiceStateUpdate) {
	a.levelManager.OnVoiceUpdate(update)
}