package payloads

import "time"

type StreamerPayload struct{
	Data []Streamer `json:"data"`
}

type Streamer struct{
	BroadcasterLanguage string `json:"broadcaster_language"`
	DisplayName string `json:"display_name"`
	GameID string `json:"game_id"`
	ID string `json:"id"`
	IsLive bool `json:"is_live"`
	ThumbnailURL string `json:"thumbnail_url"`
	Title string `json:"title"`
	StartedAt time.Time `json:"started_at"`
}
