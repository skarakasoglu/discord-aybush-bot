package models

import "time"

type DiscordEpisode struct {
	Id int
	Name string
	StartTimestamp time.Time
	EndTimestamp time.Time
}