package models

import "time"

type DiscordEpisodeExperience struct {
	DiscordMember
	CurrentLevel DiscordLevel
	NextLevel DiscordLevel
	Episode DiscordEpisode
	ExperiencePoints uint64
	ActiveVoiceMinutes int64
	LastMessageTimestamp time.Time
}
