package payloads

import "time"

type StreamChangedPayload struct{
	Data []StreamChanged `json:"data"`
}

type StreamChanged struct{
	ID string `json:"id"`
	UserID string `json:"user_id"`
	Username string `json:"user_name"`
	GameId string `json:"game_id"`
	Type string `json:"type"`
	Title string `json:"title"`
	ViewerCount int `json:"viewer_count"`
	StartedAt time.Time `json:"started_at"`
	Language string `json:"language"`
	ThumbnailUrl string `json:"thumbnail_url"`
}
