package twitch

import (
	"fmt"
	"github.com/skarakasoglu/discord-aybush-bot/configuration"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/messages"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/payloads"
	"log"
	"time"
)

var (
	BASE_API_URL = ""
	DEFAULT_API_VER = "v1"
	leaseSeconds = 864000
)

type Manager struct{
	streamer payloads.User

	oauthToken string
	userOauthToken string

	clientSecret string
	clientId string
	authorizationCode string
	redirectUri string

	apiClient *ApiClient

	userFollowsChan chan<- payloads.UserFollows
	streamChangedChan chan<- messages.StreamChanged
	running bool
}

func NewManager(streamerUsername string, clientSecret string, clientID string, authorizationCode string, redirectUri string,
	userFollowsChan chan<- payloads.UserFollows,
	streamChangedChan chan<- messages.StreamChanged, hubSecretP string, baseApiURL string) *Manager{
	hubSecret = hubSecretP
	BASE_API_URL = fmt.Sprintf("http://%v/api/twitch", baseApiURL)

	return &Manager{
		streamer: payloads.User{Login: streamerUsername},
		clientSecret: clientSecret,
		clientId: clientID,
		authorizationCode: authorizationCode,
		redirectUri: redirectUri,
		userFollowsChan: userFollowsChan,
		streamChangedChan: streamChangedChan,
		running: false,
		apiClient: NewApiClient(clientID, clientSecret, authorizationCode, redirectUri),
	}
}

func (man *Manager) IsRunning() bool {
	return man.running
}

func (man *Manager) Start() error {
	man.running = true

	man.streamer = man.apiClient.getUserInfoByUsername(man.streamer.Login)

	srv := NewServer(configuration.Manager.TwitchApi.Address, configuration.Manager.TwitchApi.Port,
		man.apiClient,
		man.userFollowsChan, man.streamChangedChan)
	go func () {
		err := srv.Start()
		if err != nil {
			log.Printf("Error on starting the server: %v", err)
		}
	}()


	time.Sleep(time.Duration(2) * time.Second)
	go func(userID int, leaseSeconds int) {
		for man.running {
			man.apiClient.subscribeToStreamChangedEvent(userID, leaseSeconds)
			man.apiClient.subscribeToUserFollowsEvent(userID, leaseSeconds)

			time.Sleep(time.Duration(leaseSeconds) * time.Second)
		}
	}(man.streamer.Id, leaseSeconds)

	return nil
}

func (man *Manager) Stop() {
	man.running = false
	man.apiClient.unsubscribeFromStreamChangedEvent(man.streamer.Id, leaseSeconds)
	man.apiClient.unsubscribeFromUserFollowsEvent(man.streamer.Id, leaseSeconds)
	// Wait for receiving unsubscribe request from twitch API.
	time.Sleep(time.Duration(5) * time.Second)
}
