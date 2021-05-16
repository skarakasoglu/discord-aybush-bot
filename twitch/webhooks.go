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

type webhookRequest struct{
	Callback string `json:"hub.callback"`
	Mode string `json:"hub.mode"`
	Topic string `json:"hub.topic"`
	LeaseSeconds int `json:"hub.lease_seconds"`
	Secret string `json:"hub.secret"`
}

func (api *ApiClient) subscribeToStreamChangedEvent(userID int, leaseSeconds int) {
	webhookReq := webhookRequest{
		Callback:     fmt.Sprintf("%v/%v/streams", BASE_API_URL, DEFAULT_API_VER),
		Mode:         "subscribe",
		Topic:        fmt.Sprintf("https://api.twitch.tv/helix/streams?user_id=%v", userID),
		LeaseSeconds: leaseSeconds,
		Secret:       hubSecret,
	}
	api.makeWebhookRequest(webhookReq)
}

func (api *ApiClient) unsubscribeFromStreamChangedEvent(userID int, leaseSeconds int) {
	webhookReq := webhookRequest{
		Callback:     fmt.Sprintf("%v/%v/streams", BASE_API_URL, DEFAULT_API_VER),
		Mode:         "unsubscribe",
		Topic:        fmt.Sprintf("https://api.twitch.tv/helix/streams?user_id=%v", userID),
		LeaseSeconds: leaseSeconds,
		Secret:       hubSecret,
	}
	api.makeWebhookRequest(webhookReq)
}

func (api *ApiClient) subscribeToUserFollowsEvent(userID int, leaseSeconds int) {
	webhookReq := webhookRequest{
		Callback:     fmt.Sprintf("%v/%v/follows", BASE_API_URL, DEFAULT_API_VER),
		Mode:         "subscribe",
		Topic:        fmt.Sprintf("https://api.twitch.tv/helix/users/follows?first=1&to_id=%v", userID),
		LeaseSeconds: leaseSeconds,
		Secret:       hubSecret,
	}
	api.makeWebhookRequest(webhookReq)
}

func (api *ApiClient) unsubscribeFromUserFollowsEvent(userID int, leaseSeconds int) {
	webhookReq := webhookRequest{
		Callback:     fmt.Sprintf("%v/%v/follows", BASE_API_URL, DEFAULT_API_VER),
		Mode:         "unsubscribe",
		Topic:        fmt.Sprintf("https://api.twitch.tv/helix/users/follows?first=1&to_id=%v", userID),
		LeaseSeconds: leaseSeconds,
		Secret:       hubSecret,
	}
	api.makeWebhookRequest(webhookReq)
}

func (api *ApiClient) makeWebhookRequest(webhookReq webhookRequest) {
	webhookURL := "https://api.twitch.tv/helix/webhooks/hub"

	reqBuffer, err := json.Marshal(webhookReq)
	if err != nil {
		log.Printf("Error on marshalling to json: %v", err)
	}

	reqBody := bytes.NewReader(reqBuffer)
	log.Printf(string(reqBuffer))

	req, err := http.NewRequest(http.MethodPost, webhookURL, reqBody)
	if err != nil {
		log.Printf("Error on making request to the end point: %v", err)
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", api.appAccessToken))
	req.Header.Set("Client-ID", api.clientId)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if resp.StatusCode != http.StatusAccepted {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error on reading response body: %v", err)
		}

		log.Printf("Error on response status: %v, body: %v", resp.StatusCode, string(body))
	}
}
