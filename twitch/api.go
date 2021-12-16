package twitch

import (
	"github.com/skarakasoglu/discord-aybush-bot/twitch/messages"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/payloads/v1"
)

type api struct{
	apiClient *ApiClient
	userFollowsChan chan<- v1.UserFollows
	streamChangedChan chan<- messages.StreamChanged
	receivedNotifications map[string]string
}