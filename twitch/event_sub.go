package twitch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// Event Types
const (
	EventType_ChannelFollow = "channel.follow"
	EventType_StreamOnline = "stream.online"
	EventType_StreamOffline = "stream.offline"
)

// Event Methods
const (
	EventMethod_webhook = "webhook"
)

const (
	SUBSCRIPTION_VERSION = "1"
	EVENT_SUB_URL = "https://api.twitch.tv/helix/eventsub/subscriptions"
)

type eventSubRequest struct{
	Type string `json:"type"`
	Version string `json:"version"`
	Condition interface{} `json:"condition"`
	Transport struct{
		Method string `json:"method"`
		Callback string `json:"callback"`
		Secret string `json:"secret"`
	} `json:"transport"`
}

func (api *ApiClient) subscribeToStreamOnlineEvent(userId string, secret string) {
	eventSubReq := eventSubRequest{
		Type:      EventType_StreamOnline,
		Version:   SUBSCRIPTION_VERSION,
		Condition: struct{
			BroadcasterUserId string `json:"broadcaster_user_id"`
		}{
			BroadcasterUserId: userId,
		},
		Transport: struct {
				Method   string `json:"method"`
				Callback string `json:"callback"`
				Secret   string `json:"secret"`
			}{
			Method:   EventMethod_webhook,
			Callback: fmt.Sprintf("%v/%v/streams/%v", BASE_SELF_API_URL, DEFAULT_SELF_API_VER, userId),
			Secret:   secret,
		},
	}
	api.makeEventSubRequest(eventSubReq)
}

func (api *ApiClient) subscribeToStreamOfflineEvent(userId string, secret string) {
	eventSubReq := eventSubRequest{
		Type:      EventType_StreamOffline,
		Version:   SUBSCRIPTION_VERSION,
		Condition: struct{
			BroadcasterUserId string `json:"broadcaster_user_id"`
		}{
			BroadcasterUserId: userId,
		},
		Transport: struct {
			Method   string `json:"method"`
			Callback string `json:"callback"`
			Secret   string `json:"secret"`
		}{
			Method:   EventMethod_webhook,
			Callback: fmt.Sprintf("%v/%v/streams/%v", BASE_SELF_API_URL, DEFAULT_SELF_API_VER, userId),
			Secret:   secret,
		},
	}
	api.makeEventSubRequest(eventSubReq)
}

func (api *ApiClient) subscribeToChannelFollowEvent(userId string, secret string) {
	eventSubReq := eventSubRequest{
		Type:      EventType_ChannelFollow,
		Version:   SUBSCRIPTION_VERSION,
		Condition: struct{
			BroadcasterUserId string `json:"broadcaster_user_id"`
		}{
			BroadcasterUserId: userId,
		},
		Transport: struct {
			Method   string `json:"method"`
			Callback string `json:"callback"`
			Secret   string `json:"secret"`
		}{
			Method:   EventMethod_webhook,
			Callback: fmt.Sprintf("%v/%v/follows/%v", BASE_SELF_API_URL, DEFAULT_SELF_API_VER, userId),
			Secret:   secret,
		},
	}
	api.makeEventSubRequest(eventSubReq)
}

func (api *ApiClient) makeEventSubRequest(eventSubReq eventSubRequest) {
	reqBuffer, err := json.Marshal(eventSubReq)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on marshalling to json: %v", err)
	}

	reqBody := bytes.NewReader(reqBuffer)

	req, err := http.NewRequest(http.MethodPost, EVENT_SUB_URL, reqBody)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on making request to the end point: %v", err)
		return
	}

	log.Printf("[TwitchEventSubAPI] Webhook subscribe/unsubscribe request: %v", string(reqBuffer))

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", api.appAccessToken))
	req.Header.Set("Client-ID", api.clientId)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if resp.StatusCode != http.StatusAccepted {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("[TwitchApiClient] Error on reading response body: %v", err)
		}

		log.Printf("[TwitchApiClient] Error on response status: %v, body: %v", resp.StatusCode, string(body))
	}
}