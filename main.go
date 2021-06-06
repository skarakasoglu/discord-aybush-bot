package main

import (
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/skarakasoglu/discord-aybush-bot/bot"
	"github.com/skarakasoglu/discord-aybush-bot/configuration"
	"github.com/skarakasoglu/discord-aybush-bot/data"
	"github.com/skarakasoglu/discord-aybush-bot/service"
	"github.com/skarakasoglu/discord-aybush-bot/shopier"
	"github.com/skarakasoglu/discord-aybush-bot/shopier/models"
	"github.com/skarakasoglu/discord-aybush-bot/streamlabs"
	"github.com/skarakasoglu/discord-aybush-bot/twitch"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/messages"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/payloads"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	configurationFileName string
	configurationFilePath string
	discordAccessToken string
	twitchClientId string
	twitchClientSecret string
	twitchRefreshToken string
	hubSecret string
	baseApiAddress string
	dbHost string
	dbPort int
	dbUsername string
	dbPassword string
	dbName string
	certFile string
	keyFile string
	streamLabsAccessToken string
	shopierUsername string
	shopierKey string
)

func init() {
	flag.StringVar(&discordAccessToken,"discord-token", "", "discord api application access token")
	flag.StringVar(&configurationFileName,"cfg-file", "config", "application configuration file name")
	flag.StringVar(&configurationFilePath, "cfg-file-path", ".", "application configuration file path")
	flag.StringVar(&twitchClientId, "twitch-client-id", "", "twitch api client id")
	flag.StringVar(&twitchClientSecret, "twitch-client-secret", "", "twitch api client secret to generate access token")
	flag.StringVar(&twitchRefreshToken, "twitch-refresh-token", "", "twitch refresh token to regenerate access token when it expires")
	flag.StringVar(&hubSecret, "hub-secret", "", "twitch webhook api secret")
	flag.StringVar(&baseApiAddress, "base-api-address", "", "twitch webhook api server address")
	flag.StringVar(&dbHost, "db-ip-address", "", "database ip address")
	flag.IntVar(&dbPort, "db-port", 0, "database port")
	flag.StringVar(&dbUsername, "db-username", "", "database login username")
	flag.StringVar(&dbPassword, "db-password", "", "database login password")
	flag.StringVar(&dbName, "db-name", "", "database name")
	flag.StringVar(&certFile, "cert-file", "", "ssl certificate file")
	flag.StringVar(&keyFile, "key-file", "", "ssl private key file")
	flag.StringVar(&streamLabsAccessToken, "streamlabs-access-token", "", "streamlabs api access token")
	flag.StringVar(&shopierUsername, "shopier-username", "", "shopier order notification api username")
	flag.StringVar(&shopierKey, "shopier-key", "", "shopier order notification api key")
	flag.Parse()

	configuration.ReadConfigurationFile(configurationFilePath, configurationFileName)
}

func main() {
	dg, err := discordgo.New(fmt.Sprintf("Bot %v", discordAccessToken))
	if err != nil {
		log.Fatalf("[AybushBot] Failed to create: %v", err)
	}

	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)
	err = dg.Open()
	if err != nil {
		log.Fatalf("[AybushBot] Failed to open websocket connection with Discord API: %v", err)
	}

	userFollowChan := make(chan payloads.UserFollows)
	streamChangedChan := make(chan messages.StreamChanged)
	shopierOrderChan := make(chan models.Order)

	db, err := data.NewDB(data.DatabaseCredentials{
		Host:         dbHost,
		Port:         dbPort,
		Username:     dbUsername,
		Password:     dbPassword,
		DatabaseName: dbName,
	}, data.PoolSettings{
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifeTime: time.Duration(30) * time.Minute,
	})
	if err != nil {
		log.Printf("[AybushBot] Error on creating db connection: %v", err)
		return
	}
	defer db.Close()

	streamLabsApi := streamlabs.NewApiClient(streamLabsAccessToken)

	discordService := service.NewDiscordService(db)
	twitchService := service.NewTwitchService(db)
	streamerUsername := "aybusee"

	aybusBot := bot.New(dg, userFollowChan, streamChangedChan, shopierOrderChan, discordService, streamLabsApi)
	aybusBot.Start()

	twitchWebhookManager := twitch.NewManager(streamerUsername, twitchClientSecret, twitchClientId, twitchRefreshToken, userFollowChan, streamChangedChan, hubSecret, twitchService, certFile, keyFile)
	err = twitchWebhookManager.Start()


	shoppierSrv := shopier.NewServer(shopierUsername, shopierKey, certFile, keyFile, shopierOrderChan)
	shoppierSrv.Start()

	log.Println("AYBUSH BOT is now running. Press CTRL + C to interrupt.")
	signalHandler := make (chan os.Signal)
	signal.Notify(signalHandler, os.Interrupt, os.Kill, syscall.SIGUSR1, syscall.SIGTERM)
	receivedSignal := <-signalHandler

	log.Printf("[AybushBot] %v signal received. Gracefully shutting down the application.", receivedSignal)
	aybusBot.Stop()
	twitchWebhookManager.Stop()

	log.Printf("[AybushBot] Application exited.")
}