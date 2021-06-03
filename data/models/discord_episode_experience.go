package models

import "time"

type DiscordEpisodeExperience struct {
	DiscordMember
	Episode DiscordEpisode
	ExperiencePoints uint64
	ActiveVoiceMinutes int64
	LastMessageTimestamp time.Time
}
