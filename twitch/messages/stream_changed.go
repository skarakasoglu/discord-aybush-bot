package messages

import "time"

type StreamChanged struct {
	UserID string
	Title string
	Username string
	GameName string
	AvatarURL string
	ThumbnailURL string
	ViewerCount int
	StartedAt time.Time
}
