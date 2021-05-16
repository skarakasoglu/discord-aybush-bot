package payloads

import "time"

type UserPayload struct{
	Data []User `json:"data"`
}

type User struct{
	Id string `json:"id"`
	Login string `json:"login"`
	DisplayName string `json:"display_name"`
	ProfileImageUrl string `json:"profile_image_url"`
	Description string `json:"description"`
	BroadcasterType string `json:"broadcaster_type"`
	ViewCount int `json:"view_count"`
	CreatedAt time.Time `json:"created_at"`
}
