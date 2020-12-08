package twitch

import (
	"encoding/json"
	"fmt"
	"github.com/skarakasoglu/discord-aybush-bot/configuration"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/messages"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/payloads"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var (
	BASE_API_URL = "http://176.53.90.209:8080/api/twitch"
	DEFAULT_API_VER = "v1"
	userID = "176613744"
	leaseSeconds = 864000
)

type Manager struct{
	oauthToken string
	clientId string
	userFollowsChan chan<- payloads.UserFollows
	streamChangedChan chan<- messages.StreamChanged
	running bool
}

func NewManager(oauthToken string, clientID string,
	userFollowsChan chan<- payloads.UserFollows,
	streamChangedChan chan<- messages.StreamChanged) *Manager{
	return &Manager{
		oauthToken: oauthToken,
		clientId: clientID,
		userFollowsChan: userFollowsChan,
		streamChangedChan: streamChangedChan,
		running: false,
	}
}

func (man *Manager) IsRunning() bool {
	return man.running
}

func (man *Manager) Start() error {
	man.running = true

	srv := NewServer(configuration.Manager.TwitchApi.Address, configuration.Manager.TwitchApi.Port,
		man,
		man.userFollowsChan, man.streamChangedChan)
	go func () {
		err := srv.Start()
		if err != nil {
			log.Printf("Error on starting the server: %v", err)
		}
	}()


	time.Sleep(time.Duration(2) * time.Second)
	go func(userID string, leaseSeconds int) {
		for man.running {
			man.subscribeToStreamChangedEvent(userID, leaseSeconds)
			man.subscribeToUserFollowsEvent(userID, leaseSeconds)

			time.Sleep(time.Duration(leaseSeconds) * time.Second)
		}
	}(userID, leaseSeconds)

	return nil
}

func (man *Manager) Stop() {
	man.running = false
	man.unsubscribeFromStreamChangedEvent(userID, leaseSeconds)
	man.unsubscribeFromUserFollowsEvent(userID, leaseSeconds)
	// Wait for receiving unsubscribe request from twitch API.
	time.Sleep(time.Duration(5) * time.Second)
}

func (man *Manager) getStreamerByUsername(username string) payloads.Streamer {
	gameReqUrl := fmt.Sprintf("https://api.twitch.tv/helix/search/channels?query=%v", username)
	resp, err := man.makeHttpGetRequest(gameReqUrl)
	if err != nil {
		log.Printf("Error on making request: %v", err)
		return payloads.Streamer{}
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error on reading response body: %v", err)
		return payloads.Streamer{}
	}

	var streamerPayload payloads.StreamerPayload
	err = json.Unmarshal(buffer, &streamerPayload)
	if err != nil {
		log.Printf("Error on unmarshalling json: %v", err)
	}

	if len(streamerPayload.Data) < 1 {
		log.Printf("No streamer found.")
		return payloads.Streamer{}
	}

	return streamerPayload.Data[0]
}

func (man *Manager) getGameById(gameID string) payloads.Game {
	gameReqUrl := fmt.Sprintf("https://api.twitch.tv/helix/games?id=%v", gameID)
	resp, err := man.makeHttpGetRequest(gameReqUrl)
	if err != nil {
		log.Printf("Error on making request: %v", err)
		return payloads.Game{}
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error on reading response body: %v", err)
		return payloads.Game{}
	}

	var gamePayload payloads.GamePayload
	err = json.Unmarshal(buffer, &gamePayload)
	if err != nil {
		log.Printf("Error on unmarshalling json: %v", err)
	}

	if len(gamePayload.Data) < 1 {
		log.Printf("No game found.")
		return payloads.Game{}
	}

	return gamePayload.Data[0]
}

func (man *Manager) makeHttpGetRequest(requestURL string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)

	if err != nil {
		log.Printf("Error on creating new request: %v", req)
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", man.oauthToken))
	req.Header.Set("Client-ID", man.clientId)

	client := &http.Client{}
	resp, err := client.Do(req)
	return resp, err
}