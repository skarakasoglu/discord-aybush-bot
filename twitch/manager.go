package twitch

import (
	"github.com/skarakasoglu/discord-aybush-bot/repository"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/messages"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/payloads"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/payloads/v1"
	"github.com/thanhpk/randstr"
	"log"
	"time"
)

var (
	BASE_SELF_API_URL    = "https://twitch.aybushbot.com/api"
	DEFAULT_SELF_API_VER = "v2"
	leaseSeconds         = 864000
	CURRENT_TWITCH_API_VER = 2
)

var broadcasterEventSecrets = map[string]map[string]string{}

type Manager struct{
	streamer payloads.User

	oauthToken string
	userOauthToken string

	clientSecret string
	clientId string
	userRefreshToken string

	certFile string
	keyFile string

	apiServer *server
	apiClient *ApiClient
	chatBot *ChatBot

	userFollowsChan chan<- v1.UserFollows
	streamChangedChan chan<- messages.StreamChanged
	running bool

	twitchRepository repository.TwitchRepository
}

func NewManager(streamerUsername string, clientSecret string, clientID string, userRefreshToken string,
	userFollowsChan chan<- v1.UserFollows,
	streamChangedChan chan<- messages.StreamChanged, hubSecretP string, twitchRepository repository.TwitchRepository,
	certFile string, keyFile string) *Manager{
	hubSecret = hubSecretP

	return &Manager{
		streamer: payloads.User{Login: streamerUsername},
		clientSecret: clientSecret,
		clientId: clientID,
		userRefreshToken: userRefreshToken,
		userFollowsChan: userFollowsChan,
		streamChangedChan: streamChangedChan,
		running: false,
		twitchRepository: twitchRepository,
		apiClient: NewApiClient(clientID, clientSecret, userRefreshToken),
		certFile: certFile,
		keyFile: keyFile,
	}
}

func (man *Manager) IsRunning() bool {
	return man.running
}

func (man *Manager) Start() error {
	man.running = true

	man.streamer = man.apiClient.getUserInfoByUsername(man.streamer.Login)

	man.chatBot = NewChatBot("aybushbot", man.apiClient.userAccessToken, man.streamer, man.apiClient, man.twitchRepository)
	man.chatBot.Start()

	man.apiServer = NewServer(man.apiClient,
		man.userFollowsChan, man.streamChangedChan, man.certFile, man.keyFile)
	go func () {
		err := man.apiServer.Start()
		if err != nil {
			log.Printf("[TwitchManager] Error on starting the server: %v", err)
		}
	}()


	man.unsubscribeFromAll()
	time.Sleep(time.Duration(2) * time.Second)

	// THE TEAM WILL BE RETRIEVED FROM DB.
	// THIS IS FOR TESTING PURPOSES.
	// IT WILL BE REMOVED IN THE NEXT CHANGE.
	broadcasters := []string{"aybusee", "Rioym", "bidik", "yoshiwou"}

	for _, broadcaster := range broadcasters {
		streamOnlineSecret := randstr.String(10)
		channelFollowSecret := randstr.String(10)
		streamOfflineSecret := randstr.String(10)
		userInfo := man.apiClient.getUserInfoByUsername(broadcaster)

		broadcasterEventSecrets[userInfo.Id] = make(map[string]string)

		broadcasterEventSecrets[userInfo.Id][EventType_StreamOnline] = streamOnlineSecret
		broadcasterEventSecrets[userInfo.Id][EventType_StreamOffline] = streamOfflineSecret
		broadcasterEventSecrets[userInfo.Id][EventType_ChannelFollow] = channelFollowSecret


		man.apiClient.subscribeToStreamOnlineEvent(userInfo.Id, streamOnlineSecret)
		man.apiClient.subscribeToChannelFollowEvent(userInfo.Id, channelFollowSecret)
	}

	return nil
}

func (man *Manager) Stop() {
	man.running = false

	man.chatBot.Stop()

	//man.apiClient.unsubscribeFromStreamChangedEvent(man.streamer.Id, leaseSeconds)
	//man.apiClient.unsubscribeFromUserFollowsEvent(man.streamer.Id, leaseSeconds)
	// Wait for receiving unsubscribe request from twitch API.
	man.unsubscribeFromAll()

	time.Sleep(time.Duration(5) * time.Second)
}

func (man *Manager) unsubscribeFromAll() {
	subscriptions := man.apiClient.getAllSubscriptions()
	for _, sub := range subscriptions {
		man.apiClient.deleteSubscription(sub)
	}
}
