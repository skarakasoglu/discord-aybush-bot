package v2

import "time"

const (
	SubscriptionStatus_Verification_Pending = "webhook_callback_verification_pending"
	SubscriptionStatus_Enabled = "enabled"

	EventSubMessageType_Notification = "notification"
	EventSubMessageType_Verification = "webhook_callback_verification"
)

type SubscriptionPayload struct{
	Total int `json:"total"`
	Data []Subscription `json:"data"`
}

type Subscription struct{
	Id string `json:"id"`
	Status string `json:"status"`
	Type string `json:"type"`
	Version string `json:"version"`
	Cost int `json:"cost"`
	Condition interface{} `json:"condition"`
	Transport struct{
		Method string `json:"method"`
		Callback string `json:"callback"`
	} `json:"transport"`
	CreatedAt time.Time `json:"created_at"`
}
