package twitch

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/messages"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/payloads"
	"io/ioutil"
	"log"
	"net/http"
)

type apiV1 struct{
	apiClient *ApiClient
	userFollowsChan chan<- payloads.UserFollows
	streamChangedChan chan<- messages.StreamChanged
	receivedNotifications map[string]string
}

func NewApiV1(apiClient *ApiClient, userFollowsChan chan<- payloads.UserFollows,
	streamChangedChan chan<- messages.StreamChanged) *apiV1{
	return &apiV1{
		apiClient: apiClient,
		userFollowsChan: userFollowsChan,
		streamChangedChan: streamChangedChan,
		receivedNotifications: make(map[string]string),
	}
}

func (api *apiV1) onSubscriptionValidated(ctx *gin.Context) {
	denyReason := ctx.Query("hub.reason")

	if denyReason != "" {
		log.Printf("[TwitchWebhookAPI] Subscription denied: %v", denyReason)
	} else {
		hubChallengeString := ctx.Query("hub.challenge")
		topic := ctx.Query("hub.topic")
		subscriptionSeconds := ctx.Query("hub.lease_seconds")
		mode := ctx.Query("hub.mode")

		log.Printf("[TwitchWebhookAPI] %v validated to %v %v seconds.", mode, topic, subscriptionSeconds)
		ctx.Header("Content-Type", "text/plain")
		ctx.String(http.StatusOK, hubChallengeString)
	}
}

func (api *apiV1) onStreamChanged(ctx *gin.Context) {
	buffer, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Printf("[TwitchWebhookAPI] Error on reading request body: %v", err)
		return
	}

	var streamChangePayload payloads.StreamChangedPayload
	err = json.Unmarshal(buffer, &streamChangePayload)
	if err != nil {
		log.Printf("[TwitchWebhookAPI] Error on binding to request json: %v", err)
		return
	}

	notificationId := ctx.GetHeader("Twitch-Notification-Id")
	_, ok := api.receivedNotifications[notificationId]
	if ok {
		log.Printf("[TwitchWebhookAPI] Duplicate streamChanged notification received from twitch: %v", notificationId)
	} else {
		var streamChanged messages.StreamChanged

		if len(streamChangePayload.Data) < 1 {
			streamChanged.UserID = "0"

			userId := ctx.Param("userId")
			streamer := api.apiClient.getUserInfoByUserId(userId)

			streamChanged = messages.StreamChanged{
				Username:     streamer.Login,
			}
		} else {
			signature := ctx.GetHeader("X-Hub-Signature")
			valid, signatureShouldBe := api.validateSignature(signature, buffer)
			if !valid{
				log.Printf("[TwitchWebhookAPI] The payload signature is not valid. Payload: %+v Unauthenticated request, signature: %v, signatureShouldBe: %v",
					streamChangePayload.Data[0], signature, signatureShouldBe)
				ctx.String(http.StatusOK, "")
				return
			}

			api.receivedNotifications[notificationId] = notificationId

			streamChangeInfo := streamChangePayload.Data[0]
			log.Printf("[TwitchWebhookAPI] Notification id: %v stream changed end point called: %+v", notificationId, streamChangeInfo)

			streamer := api.apiClient.getUserInfoByUsername(streamChangeInfo.Username)
			game := api.apiClient.getGameById(streamChangeInfo.GameId)

			streamChanged = messages.StreamChanged{
				UserID:       streamChangeInfo.UserID,
				Title:        streamChangeInfo.Title,
				Username:     streamChangeInfo.Username,
				GameName:     game.Name,
				AvatarURL:    streamer.ProfileImageUrl,
				ThumbnailURL: streamChangeInfo.ThumbnailUrl,
				ViewerCount:  streamChangeInfo.ViewerCount,
				StartedAt:    streamChangeInfo.StartedAt.Local(),
			}
		}

		api.streamChangedChan <- streamChanged
	}

	ctx.String(http.StatusOK, "")
}

func (api *apiV1) onUserFollows(ctx *gin.Context) {
	buffer, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Printf("[TwitchWebhookAPI] Error on reading request body: %v", err)
		return
	}

	var followPayload payloads.UserFollowsPayload
	err = json.Unmarshal(buffer, &followPayload)
	if err != nil {
		log.Printf("[TwitchWebhookAPI] Error on binding to request json: %v", err)
		return
	}

	signature := ctx.GetHeader("X-Hub-Signature")
	valid, signatureShouldBe := api.validateSignature(signature, buffer)
	if !valid{
		log.Printf("[TwitchWebhookAPI] The payload signature is not valid. Unauthenticated request, signature: %v, signatureShouldBe: %v",
			signature, signatureShouldBe)
		ctx.String(http.StatusOK, "")
		return
	}

	userId := ctx.Param("userId")

	notificationId := ctx.GetHeader("Twitch-Notification-Id")
	_, ok := api.receivedNotifications[notificationId]
	if ok {
		log.Printf("[TwitchWebhookAPI] UserId: %v, Duplicate userFollows notification received from twitch: %v", userId, notificationId)
	} else {
		api.receivedNotifications[notificationId] = notificationId
		if len(followPayload.Data) < 1 {
			log.Printf("[TwitchWebhookAPI] User follow end point called but no data found.")
		} else {
			followInfo := followPayload.Data[0]
			log.Printf("[TwitchWebhookAPI] NotificationId: %v, UserId: %v User follows end point called: %v", notificationId, userId, followInfo)
			api.userFollowsChan <- followInfo
		}
	}

	ctx.String(http.StatusOK, "")
}

func (api *apiV1) validateSignature(signature string, payload []byte) (bool, string) {
	h := hmac.New(sha256.New, []byte(hubSecret))
	h.Write(payload)
	signatureShouldBe := fmt.Sprintf("sha256=%v", hex.EncodeToString(h.Sum(nil)))

	return signatureShouldBe == signature, signatureShouldBe
}