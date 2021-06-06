package shopier

import (
	"github.com/gin-gonic/gin"
	"github.com/skarakasoglu/discord-aybush-bot/shopier/models"
	"log"
)

type server struct {
	username string
	key string

	certFile string
	keyFile string

	orderNotifierChan chan<- models.Order
}

func NewServer(username string, key string, certFile string, keyFile string, orderNotifierChan chan<- models.Order) *server{
	return &server{
		username: username,
		key:      key,
		certFile: certFile,
		keyFile:  keyFile,
		orderNotifierChan: orderNotifierChan,
	}
}

func (srv *server) Start() {
	router := gin.Default()

	router.Static("/alerts", "./alerts")
	router.StaticFile("/images/venom.png", "./images/venom.png")

	apiv1 := NewApiV1(srv.username, srv.key, srv.orderNotifierChan)

	api := router.Group("/api")

	v1 := api.Group("v1")
	{
		v1.POST("/orders", apiv1.onOrderNotification)
	}

	err := router.RunTLS(":444", srv.certFile, srv.keyFile)
	if err != nil {
		log.Printf("[ShoppierAPI] Error on running the router: %v", err)
		return
	}
}