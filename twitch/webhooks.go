package twitch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	hubSecret string
)

// WEBHOOKS DEPRECATED BY TWITCH

type webhookRequest struct{
	Callback string `json:"hub.callback"`
	Mode string `json:"hub.mode"`
	Topic string `json:"hub.topic"`
	LeaseSeconds int `json:"hub.lease_seconds"`
	Secret string `json:"hub.secret"`
}

func (api *ApiClient) subscribeToStreamChangedEvent(userID string, leaseSeconds int) {
	webhookReq := webhookRequest{
		Callback:     fmt.Sprintf("%v/%v/streams/%v", BASE_SELF_API_URL, DEFAULT_SELF_API_VER, userID),
		Mode:         "subscribe",
		Topic:        fmt.Sprintf("%v/%v/%v?user_id=%v", BASE_API_URL, API_VERSION, STREAMS_ENDPOINT, userID),
		LeaseSeconds: leaseSeconds,
		Secret:       hubSecret,
	}
	api.makeWebhookRequest(webhookReq)
}

func (api *ApiClient) unsubscribeFromStreamChangedEvent(userID string, leaseSeconds int) {
	webhookReq := webhookRequest{
		Callback:     fmt.Sprintf("%v/%v/streams/%v", BASE_SELF_API_URL, DEFAULT_SELF_API_VER, userID),
		Mode:         "unsubscribe",
		Topic:        fmt.Sprintf("%v/%v/%v?user_id=%v", BASE_API_URL, API_VERSION, STREAMS_ENDPOINT, userID),
		LeaseSeconds: leaseSeconds,
		Secret:       hubSecret,
	}
	api.makeWebhookRequest(webhookReq)
}

func (api *ApiClient) subscribeToUserFollowsEvent(userID string, leaseSeconds int) {
	webhookReq := webhookRequest{
		Callback:     fmt.Sprintf("%v/%v/follows/%v", BASE_SELF_API_URL, DEFAULT_SELF_API_VER, userID),
		Mode:         "subscribe",
		Topic:        fmt.Sprintf("%v/%v/%v?first=1&to_id=%v", BASE_API_URL, API_VERSION, FOLLOWS_ENDPOINT, userID),
		LeaseSeconds: leaseSeconds,
		Secret:       hubSecret,
	}
	api.makeWebhookRequest(webhookReq)
}

func (api *ApiClient) unsubscribeFromUserFollowsEvent(userID string, leaseSeconds int) {
	webhookReq := webhookRequest{
		Callback:     fmt.Sprintf("%v/%v/follows/%v", BASE_SELF_API_URL, DEFAULT_SELF_API_VER, userID),
		Mode:         "unsubscribe",
		Topic:        fmt.Sprintf("%v/%v/%v?first=1&to_id=%v", BASE_API_URL, API_VERSION, FOLLOWS_ENDPOINT, userID),
		LeaseSeconds: leaseSeconds,
		Secret:       hubSecret,
	}
	api.makeWebhookRequest(webhookReq)
}

func (api *ApiClient) makeWebhookRequest(webhookReq webhookRequest) {
	webhookURL := "https://api.twitch.tv/helix/webhooks/hub"

	reqBuffer, err := json.Marshal(webhookReq)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on marshalling to json: %v", err)
	}

	reqBody := bytes.NewReader(reqBuffer)

	req, err := http.NewRequest(http.MethodPost, webhookURL, reqBody)
	if err != nil {
		log.Printf("[TwitchApiClient] Error on making request to the end point: %v", err)
		return
	}

	log.Printf("[TwitchWebhookAPI] Webhook subscribe/unsubscribe request: %v", string(reqBuffer))

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
