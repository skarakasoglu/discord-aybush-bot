package twitch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type webhookRequest struct{
	Callback string `json:"hub.callback"`
	Mode string `json:"hub.mode"`
	Topic string `json:"hub.topic"`
	LeaseSeconds int `json:"hub.lease_seconds"`
	Secret string `json:"hub.secret"`
}

func (man *Manager) subscribeToStreamChangedEvent(userID string, leaseSeconds int) {
	webhookReq := webhookRequest{
		Callback:     fmt.Sprintf("%v/%v/streams", BASE_API_URL, DEFAULT_API_VER),
		Mode:         "subscribe",
		Topic:        fmt.Sprintf("https://api.twitch.tv/helix/streams?user_id=%v", userID),
		LeaseSeconds: leaseSeconds,
		Secret:       "aybush",
	}
	man.makeWebhookRequest(webhookReq)
}

func (man *Manager) unsubscribeFromStreamChangedEvent(userID string, leaseSeconds int) {
	webhookReq := webhookRequest{
		Callback:     fmt.Sprintf("%v/%v/streams", BASE_API_URL, DEFAULT_API_VER),
		Mode:         "unsubscribe",
		Topic:        fmt.Sprintf("https://api.twitch.tv/helix/streams?user_id=%v", userID),
		LeaseSeconds: leaseSeconds,
		Secret:       "aybush",
	}
	man.makeWebhookRequest(webhookReq)
}

func (man *Manager) subscribeToUserFollowsEvent(userID string, leaseSeconds int) {
	webhookReq := webhookRequest{
		Callback:     fmt.Sprintf("%v/%v/follows", BASE_API_URL, DEFAULT_API_VER),
		Mode:         "subscribe",
		Topic:        fmt.Sprintf("https://api.twitch.tv/helix/users/follows?first=1&to_id=%v", userID),
		LeaseSeconds: leaseSeconds,
		Secret:       "aybush",
	}
	man.makeWebhookRequest(webhookReq)
}

func (man *Manager) unsubscribeFromUserFollowsEvent(userID string, leaseSeconds int) {
	webhookReq := webhookRequest{
		Callback:     fmt.Sprintf("%v/%v/follows", BASE_API_URL, DEFAULT_API_VER),
		Mode:         "unsubscribe",
		Topic:        fmt.Sprintf("https://api.twitch.tv/helix/users/follows?first=1&to_id=%v", userID),
		LeaseSeconds: leaseSeconds,
		Secret:       "aybush",
	}
	man.makeWebhookRequest(webhookReq)
}

func (man *Manager) makeWebhookRequest(webhookReq webhookRequest) {
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

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", man.oauthToken))
	req.Header.Set("Client-ID", man.clientId)
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
