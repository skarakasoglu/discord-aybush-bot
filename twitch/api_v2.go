package twitch

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/messages"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/payloads/v1"
	v2 "github.com/skarakasoglu/discord-aybush-bot/twitch/payloads/v2"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type apiV2 struct{
	api
}


func NewApiV2(apiClient *ApiClient, userFollowsChan chan<- v1.UserFollows,
	streamChangedChan chan<- messages.StreamChanged) *apiV2{
	return &apiV2{
		api: api{
			apiClient: apiClient,
			userFollowsChan: userFollowsChan,
			streamChangedChan: streamChangedChan,
			receivedNotifications: make(map[string]string),
		},
	}
}

func (api *apiV2) onStreamChanged(ctx *gin.Context) {
	userId := ctx.Param("userId")
	eventSecrets, ok := broadcasterEventSecrets[userId]
	if !ok {
		log.Printf("[TwitchEventSubAPI] Cannot find broadcaster: %v", userId)
		return
	}

	subscriptionType := ctx.GetHeader("Twitch-Eventsub-Subscription-Type")
	secret, ok := eventSecrets[subscriptionType]
	if !ok {
		log.Printf("[TwitchEventSubAPI] Cannot find event secret for type: %v", subscriptionType)
		return
	}

	buffer, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Printf("[TwitchEventSubAPI] Error on reading request body: %v", err)
		return
	}

	var data map[string]interface{}
	err = json.Unmarshal(buffer, &data)
	if err != nil {
		log.Printf("[TwitchEventSubAPI] Error on unmarshalling request body: %v", err)
		return
	}

	notificationId := ctx.GetHeader("Twitch-Eventsub-Message-Id")
	_, ok = api.receivedNotifications[notificationId]
	if ok {
		log.Printf("[TwitchEventSubAPI] Duplicate streamChanged notification received from twitch: %v", notificationId)
	} else {
		api.receivedNotifications[notificationId] = notificationId
		timestampStr := ctx.GetHeader("Twitch-Eventsub-Message-Timestamp")
		timestamp, err := time.Parse(time.RFC3339Nano, timestampStr)
		if err != nil {
			log.Printf("[TwitchEventSubAPI] Error on parsing timestamp: %v", timestampStr)
			ctx.String(http.StatusBadRequest, "")
			return
		}

		timeElapsed := time.Now().Sub(timestamp)
		if timeElapsed.Minutes() >= 10 {
			log.Printf("[TwitchEventSubAPI] old notification message received: %v", timestampStr)
			ctx.String(http.StatusOK, "")
			return
		}

		signature := ctx.GetHeader("Twitch-Eventsub-Message-Signature")
		valid, signatureShouldBe := api.validateSignature(notificationId, timestampStr, string(buffer), secret, signature)
		if !valid{
			log.Printf("[TwitchEventSubAPI] The payload signature is not valid. Unauthenticated request, signature: %v, signatureShouldBe: %v",
				signature, signatureShouldBe)
			ctx.String(http.StatusOK, "")
			return
		}

		messageType := ctx.GetHeader("Twitch-Eventsub-Message-Type")
		if messageType == v2.EventSubMessageType_Verification {
			challenge, ok := data["challenge"].(string)
			if !ok {
				log.Printf("[TwitchEventSubAPI] error on obtaining challenge string from request: %v", data)
				return
			}

			log.Printf("[TwitchEventSubAPI] %v subscription verified successfully.", subscriptionType)
			ctx.String(http.StatusOK, challenge)
			return
		} else {
			event, ok := data["event"]
			if !ok {
				log.Printf("[TwitchEventSubAPI] User follow end point called but no data found.")
			} else {
				var streamChanged messages.StreamChanged

				if subscriptionType == EventType_StreamOnline {
					jsonBuf, err := json.Marshal(event)
					if err != nil {
						log.Printf("[TwitchEventSubAPI] Error on marshalling event JSON. %+v", data)
						return
					}

					var streamOnline v2.StreamOnline
					err = json.Unmarshal(jsonBuf, &streamOnline)
					if err != nil {
						log.Printf("[TwitchEventSubAPI] Error on unmarshalling event json: %v", err)
						return
					}

					log.Printf("[TwitchEventSubAPI] Notification id: %v stream changed end point called: %+v", notificationId, streamOnline)

					streamer := api.apiClient.getUserInfoByUsername(streamOnline.BroadcasterUserName)
					channelInfo := api.apiClient.getChannelInfoByBroadcasterId(streamOnline.BroadcasterUserId)

					streamChanged = messages.StreamChanged{
						StreamChangeType: messages.StreamChangeType_Live,
						UserID:       streamer.Id,
						Version: 2,
						Title:        channelInfo.Title,
						Username:     streamer.DisplayName,
						GameName:     channelInfo.GameName,
						AvatarURL:    streamer.ProfileImageUrl,
						ThumbnailURL: streamer.ProfileImageUrl,
						ViewerCount:  streamer.ViewCount,
						StartedAt:    streamOnline.StartedAt.Local(),
					}
				} else if subscriptionType == EventType_StreamOffline {
					jsonBuf, err := json.Marshal(event)
					if err != nil {
						log.Printf("[TwitchEventSubAPI] Error on marshalling event JSON. %+v", data)
						return
					}

					var streamOffline v2.StreamOffline
					err = json.Unmarshal(jsonBuf, &streamOffline)
					if err != nil {
						log.Printf("[TwitchEventSubAPI] Error on unmarshalling event json: %v", err)
						return
					}

					log.Printf("[TwitchEventSubAPI] Notification id: %v stream changed end point called: %+v", notificationId, streamOffline)

					streamChanged = messages.StreamChanged{
						StreamChangeType: messages.StreamChangeType_Offline,
						UserID:       streamOffline.BroadcasterUserId,
						Version: 2,
						Username:     streamOffline.BroadcasterUserName,
					}
				}
				api.streamChangedChan <- streamChanged
			}
		}
	}

	ctx.String(http.StatusOK, "")
}

func (api *apiV2) onUserFollows(ctx *gin.Context) {
	userId := ctx.Param("userId")
	eventSecrets, ok := broadcasterEventSecrets[userId]
	if !ok {
		log.Printf("[TwitchEventSubAPI] Cannot find broadcaster: %v", userId)
		return
	}

	subscriptionType := ctx.GetHeader("Twitch-Eventsub-Subscription-Type")
	secret, ok := eventSecrets[subscriptionType]
	if !ok {
		log.Printf("[TwitchEventSubAPI] Cannot find event secret for type: %v", subscriptionType)
		return
	}

	buffer, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Printf("[TwitchEventSubAPI] Error on reading request body: %v", err)
		return
	}

	var data map[string]interface{}
	err = json.Unmarshal(buffer, &data)
	if err != nil {
		log.Printf("[TwitchEventSubAPI] Error on unmarshalling request body: %v", err)
		return
	}

	notificationId := ctx.GetHeader("Twitch-Eventsub-Message-Id")
	_, ok = api.receivedNotifications[notificationId]
	if ok {
		log.Printf("[TwitchEventSubAPI] UserId: %v, Duplicate userFollows notification received from twitch: %v", userId, notificationId)
	} else {
		api.receivedNotifications[notificationId] = notificationId

		timestampStr := ctx.GetHeader("Twitch-Eventsub-Message-Timestamp")
		timestamp, err := time.Parse(time.RFC3339Nano, timestampStr)
		if err != nil {
			log.Printf("[TwitchEventSubAPI] Error on parsing timestamp: %v", timestampStr)
			ctx.String(http.StatusBadRequest, "")
			return
		}

		timeElapsed := time.Now().Sub(timestamp)
		if timeElapsed.Minutes() >= 10 {
			log.Printf("[TwitchEventSubAPI] old notification message received: %v", timestampStr)
			ctx.String(http.StatusOK, "")
			return
		}

		signature := ctx.GetHeader("Twitch-Eventsub-Message-Signature")
		valid, signatureShouldBe := api.validateSignature(notificationId, timestampStr, string(buffer), secret, signature)
		if !valid{
			log.Printf("[TwitchEventSubAPI] The payload signature is not valid. Unauthenticated request, signature: %v, signatureShouldBe: %v",
				signature, signatureShouldBe)
			ctx.String(http.StatusOK, "")
			return
		}

		messageType := ctx.GetHeader("Twitch-Eventsub-Message-Type")
		if messageType == v2.EventSubMessageType_Verification {
			challenge, ok := data["challenge"].(string)
			if !ok {
				log.Printf("[TwitchEventSubAPI] error on obtaining challenge string from request: %v", data)
				return
			}

			log.Printf("[TwitchEventSubAPI] %v subscription verified successfully.", subscriptionType)
			ctx.String(http.StatusOK, challenge)
			return
		} else {
			event, ok := data["event"]
			if !ok {
				log.Printf("[TwitchEventSubAPI] User follow end point called but no data found. %+v", data)
			} else {
				jsonBuf, err := json.Marshal(event)
				if err != nil {
					log.Printf("[TwitchEventSubAPI] Error on marshalling event JSON. %+v", data)
					return
				}

				var channelFollow v2.ChannelFollow
				err = json.Unmarshal(jsonBuf, &channelFollow)
				if err != nil {
					log.Printf("[TwitchEventSubAPI] Error on unmarshalling event json: %v", err)
					return
				}

				log.Printf("[TwitchEventSubAPI] NotificationId: %v, UserId: %v Channel follow end point called: %v", notificationId, userId, channelFollow)
				api.userFollowsChan <- v1.UserFollows{
					FromID:     channelFollow.UserId,
					FromName:   channelFollow.UserName,
					ToID:       channelFollow.BroadcasterUserId,
					ToName:     channelFollow.BroadcasterUserName,
					FollowedAt: channelFollow.FollowedAt,
				}
			}
		}
	}

	ctx.String(http.StatusOK, "")
}

func (*apiV2) validateSignature(messageId string, timestamp string, requestBody string, secret string, signature string) (bool, string) {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(fmt.Sprintf("%v%v%v", messageId, timestamp, requestBody)))
	signatureShouldBe := fmt.Sprintf("sha256=%v", hex.EncodeToString(h.Sum(nil)))

	return signatureShouldBe == signature, signatureShouldBe
}