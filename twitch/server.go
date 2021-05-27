package twitch

import (
	"github.com/gin-gonic/gin"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/messages"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/payloads"
	"log"
)

type server struct{
	apiClient *ApiClient
	userFollowsChan chan<- payloads.UserFollows
	streamChangedChan chan<- messages.StreamChanged

	certFile string
	keyFile string
}

type api interface{
	onSubscriptionValidated(ctx *gin.Context)
	onStreamChanged(ctx *gin.Context)
	onUserFollows(ctx *gin.Context)
}

func NewServer(apiClient *ApiClient,
	userFollowsChan chan<- payloads.UserFollows,
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

	apiv1 := NewApiV1(srv.apiClient, srv.userFollowsChan, srv.streamChangedChan)

	twitchApi := router.Group("/api")

	v1 := twitchApi.Group("/v1")
	{
		v1.GET("/streams/:userId", apiv1.onSubscriptionValidated)
		v1.GET("/follows/:userId", apiv1.onSubscriptionValidated)
		v1.POST("/streams/:userId", apiv1.onStreamChanged)
		v1.POST("/follows/:userId", apiv1.onUserFollows)
	}

	err := router.RunTLS(":443", srv.certFile, srv.keyFile)
	if err != nil {
		log.Printf("Error on running the router: %v", err)
		return err
	}

	return nil
}
