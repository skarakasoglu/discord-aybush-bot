package v2

import "time"

type StreamOnline struct {
	Id string `json:"id"`
	BroadcasterUserId string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName string `json:"broadcaster_user_name"`
	Type string `json:"type"`
	StartedAt time.Time `json:"started_at"`
}