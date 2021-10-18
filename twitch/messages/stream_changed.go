package messages

import "time"

type StreamChangeType string

const (
	StreamChangeType_Live StreamChangeType = "live"
	StreamChangeType_Offline StreamChangeType = "offline"
)

type StreamChanged struct {
	StreamChangeType StreamChangeType
	Version int
	UserID string
	Title string
	Username string
	GameName string
	AvatarURL string
	ThumbnailURL string
	ViewerCount int
	StartedAt time.Time
}
