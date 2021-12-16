package twitch

import (
	"github.com/gin-gonic/gin"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/messages"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/payloads/v1"
	"log"
)

type server struct{
	apiClient *ApiClient
	userFollowsChan chan<- v1.UserFollows
	streamChangedChan chan<- messages.StreamChanged

	certFile string
	keyFile string
}

func NewServer(apiClient *ApiClient,
	userFollowsChan chan<- v1.UserFollows,
	streamChanged chan<- messages.StreamChanged,
	certFile string, keyFile string) *server{
	return &server{
		apiClient: apiClient,
		userFollowsChan: userFollowsChan,
		streamChangedChan: streamChanged,
		certFile: certFile,
		keyFile: keyFile,
	}
}

func (srv *server) Start() error {
	router := gin.Default()

	/*
	--------- API v1 DEPRECATED BY TWITCH ----------
	apiv1 := NewApiV1(srv.apiClient, srv.userFollowsChan, srv.streamChangedChan)

	twitchApi := router.Group("/api")

	v1 := twitchApi.Group("/v1")
	{
		v1.GET("/streams/:userId", apiv1.onSubscriptionValidated)
		v1.GET("/follows/:userId", apiv1.onSubscriptionValidated)
		v1.POST("/streams/:userId", apiv1.onStreamChanged)
		v1.POST("/follows/:userId", apiv1.onUserFollows)
	}

	 */
	apiv2 := NewApiV2(srv.apiClient, srv.userFollowsChan, srv.streamChangedChan)
	twitchApi := router.Group("/api")

	v2 := twitchApi.Group("/v2")
	{
		v2.POST("/streams/:userId", apiv2.onStreamChanged)
		v2.POST("/follows/:userId", apiv2.onUserFollows)
	}

	err := router.RunTLS(":443", srv.certFile, srv.keyFile)
	if err != nil {
		log.Printf("Error on running the router: %v", err)
		return err
	}

	return nil
}
