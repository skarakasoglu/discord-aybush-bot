package twitch

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/messages"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/payloads"
	"log"
)

type server struct{
	address string
	port int
	apiClient *ApiClient
	userFollowsChan chan<- payloads.UserFollows
	streamChangedChan chan<- messages.StreamChanged
}

type api interface{
	onSubscriptionValidated(ctx *gin.Context)
	onStreamChanged(ctx *gin.Context)
	onUserFollows(ctx *gin.Context)
}

func NewServer(address string, port int,
	apiClient *ApiClient,
	userFollowsChan chan<- payloads.UserFollows,
	streamChanged chan<- messages.StreamChanged) *server{
	return &server{
		address: address,
		port: port,
		apiClient: apiClient,
		userFollowsChan: userFollowsChan,
		streamChangedChan: streamChanged,
	}
}

func (srv *server) Start() error {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	apiv1 := NewApiV1(srv.apiClient, srv.userFollowsChan, srv.streamChangedChan)

	twitchApi := router.Group("/api/twitch")

	v1 := twitchApi.Group("/v1")
	{
		v1.GET("/streams", apiv1.onSubscriptionValidated)
		v1.GET("/follows", apiv1.onSubscriptionValidated)
		v1.POST("/streams", apiv1.onStreamChanged)
		v1.POST("/follows", apiv1.onUserFollows)
	}

	err := router.Run(fmt.Sprintf("%v:%v", srv.address, srv.port))
	if err != nil {
		log.Printf("Error on running the router: %v", err)
		return err
	}

	return nil
}
