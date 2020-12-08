package twitch

import (
	"github.com/gin-gonic/gin"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/messages"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/payloads"
	"log"
	"net/http"
)

type apiV1 struct{
	manager *Manager
	userFollowsChan chan<- payloads.UserFollows
	streamChangedChan chan<- messages.StreamChanged
	receivedNotifications map[string]string
}

func NewApiV1(manager *Manager, userFollowsChan chan<- payloads.UserFollows,
	streamChangedChan chan<- messages.StreamChanged) *apiV1{
	return &apiV1{
		manager: manager,
		userFollowsChan: userFollowsChan,
		streamChangedChan: streamChangedChan,
		receivedNotifications: make(map[string]string),
	}
}

func (api *apiV1) onSubscriptionValidated(ctx *gin.Context) {
	denyReason := ctx.Query("hub.reason")

	if denyReason != "" {
		log.Printf("Subscription denied: %v", denyReason)
	} else {
		hubChallengeString := ctx.Query("hub.challenge")
		topic := ctx.Query("hub.topic")
		subscriptionSeconds := ctx.Query("hub.lease_seconds")
		mode := ctx.Query("hub.mode")

		log.Printf("%v validated to %v %v seconds.", mode, topic, subscriptionSeconds)
		ctx.Header("Content-Type", "text/plain")
		ctx.String(http.StatusOK, hubChallengeString)
	}
}

func (api *apiV1) onStreamChanged(ctx *gin.Context) {
	var streamChangePayload payloads.StreamChangedPayload
	err := ctx.BindJSON(&streamChangePayload)
	if err != nil {
		log.Printf("Error on binding to request json: %v", err)
		return
	}

	notificationId := ctx.GetHeader("Twitch-Notification-Id")
	_, ok := api.receivedNotifications[notificationId]
	if ok {
		log.Printf("Duplicate streamChanged notification received from twitch: %v", notificationId)
	} else {
		api.receivedNotifications[notificationId] = notificationId
		if len(streamChangePayload.Data) < 1 {
			log.Printf("Aybusee went ofline.")
		} else {
			streamChangeInfo := streamChangePayload.Data[0]
			log.Printf("Stream changed end point called: %v", streamChangeInfo)

			streamer := api.manager.getStreamerByUsername(streamChangeInfo.Username)
			game := api.manager.getGameById(streamChangeInfo.GameId)

			streamChanged := messages.StreamChanged{
				UserID:       streamChangeInfo.UserID,
				Title:        streamChangeInfo.Title,
				Username:     streamChangeInfo.Username,
				GameName:     game.Name,
				AvatarURL:    streamer.ThumbnailURL,
				ThumbnailURL: streamChangeInfo.ThumbnailUrl,
				ViewerCount:  streamChangeInfo.ViewerCount,
				StartedAt:    streamChangeInfo.StartedAt.Local(),
			}

			api.streamChangedChan <- streamChanged
		}
	}

	ctx.String(http.StatusOK, "")
}

func (api *apiV1) onUserFollows(ctx *gin.Context) {
	var followPayload payloads.UserFollowsPayload
	err := ctx.BindJSON(&followPayload)
	if err != nil {
		log.Printf("Error on binding to request json: %v", err)
		return
	}

	notificationId := ctx.GetHeader("Twitch-Notification-Id")
	_, ok := api.receivedNotifications[notificationId]
	if ok {
		log.Printf("Duplicate userFollows notification received from twitch: %v", notificationId)
	} else {
		api.receivedNotifications[notificationId] = notificationId
		if len(followPayload.Data) < 1 {
			log.Printf("User follow end point called but no data found.")
		} else {
			followInfo := followPayload.Data[0]
			log.Printf("User follows end point called: %v", followInfo)
			api.userFollowsChan <- followInfo
		}
	}

	ctx.String(http.StatusOK, "")
}