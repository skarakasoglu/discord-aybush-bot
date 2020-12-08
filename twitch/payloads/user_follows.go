package payloads

import "time"

type UserFollowsPayload struct{
	Data []UserFollows `json:"data"`
}

type UserFollows struct{
	FromID string `json:"from_id"`
	FromName string `json:"from_name"`
	ToID string `json:"to_id"`
	ToName string `json:"to_name"`
	FollowedAt time.Time `json:"followed_at"`
}
