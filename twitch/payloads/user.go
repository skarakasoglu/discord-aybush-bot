package payloads

import "time"

type UserPayload struct{
	Data []User `json:"data"`
}

type User struct {
	Id string `json:"id"`
	Login string `json:"login"`
	DisplayName string `json:"display_name"`
	Type string `json:"type"`
	BroadcasterType string `json:"broadcaster_type"`
	Description string `json:"description"`
	ProfileImageUrl string `json:"profile_image_url"`
	OfflineImageUrl string `json:"offline_image_url"`
	ViewCount int `json:"view_count"`
	CreatedAt time.Time `json:"created_at"`
}
