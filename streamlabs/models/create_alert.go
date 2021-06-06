package models

type AlertType uint8

const (
	AlertType_Donation AlertType = iota
	AlertType_Follow
	AlertType_Host
	AlertType_Subscription
)

var alertTypes = map[AlertType]string{
	AlertType_Donation: "donation",
	AlertType_Follow: "follow",
	AlertType_Host: "host",
	AlertType_Subscription: "subscription",
}

func (a AlertType) String() string{
	str, ok := alertTypes[a]
	if !ok {
		return ""
	}

	return str
}

type CreateAlert struct {
	Type AlertType
	ImageHref string
	SoundHref string
	Message string
	UserMessage string
	Duration int
	SpecialTextColor string
}

type CreateAlertResponse struct{
	Success bool `json:"success"`
}
