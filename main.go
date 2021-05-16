package main

import (
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/skarakasoglu/discord-aybush-bot/bot"
	"github.com/skarakasoglu/discord-aybush-bot/configuration"
	"github.com/skarakasoglu/discord-aybush-bot/data"
	"github.com/skarakasoglu/discord-aybush-bot/data/psql"
	"github.com/skarakasoglu/discord-aybush-bot/twitch"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/messages"
	"github.com/skarakasoglu/discord-aybush-bot/twitch/payloads"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	configurationFileName string
	configurationFilePath string
	discordAccessToken string
	twitchClientId string
	twitchClientSecret string
	authorizationCode string
	hubSecret string
	baseApiAddress string
	dbHost string
	dbPort int
	dbUsername string
	dbPassword string
	dbName string
)

func init() {
	flag.StringVar(&discordAccessToken,"discord-token", "", "discord api application access token")
	flag.StringVar(&configurationFileName,"cfg-file", "config", "application configuration file name")
	flag.StringVar(&configurationFilePath, "cfg-file-path", ".", "application configuration file path")
	flag.StringVar(&twitchClientId, "twitch-client-id", "", "twitch api client id")
	flag.StringVar(&twitchClientSecret, "twitch-token", "", "twitch api client secret to generate access token")
	flag.StringVar(&authorizationCode, "authorization-code", "", "twitch api authorization code to generate user access token")
	flag.StringVar(&hubSecret, "hub-secret", "", "twitch webhook api secret")
	flag.StringVar(&baseApiAddress, "base-api-address", "", "twitch webhook api server address")
	flag.StringVar(&dbHost, "db-ip-address", "", "database ip address")
	flag.IntVar(&dbPort, "db-port", 0, "database port")
	flag.StringVar(&dbUsername, "db-username", "", "database login username")
	flag.StringVar(&dbPassword, "db-password", "", "database login password")
	flag.StringVar(&dbName, "db-name", "", "database name")

	configuration.ReadConfigurationFile(configurationFilePath, configurationFileName)
}

func main() {
	dg, err := discordgo.New(fmt.Sprintf("Bot %v", discordAccessToken))
	if err != nil {
		log.Fatalf("Failed to create: %v", err)
	}

	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)
	err = dg.Open()
	if err != nil {
		log.Fatalf("Failed to open websocket connection with Discord API: %v", err)
	}

	userFollowChan := make(chan payloads.UserFollows)
	streamChangedChan := make(chan messages.StreamChanged)

	db, err := data.NewDB(data.DatabaseCredentials{
		Host:         dbHost,
		Port:         dbPort,
		Username:     dbUsername,
		Password:     dbPassword,
		DatabaseName: dbName,
	})
	if err != nil {
		log.Printf("Error on creating db connection: %v", err)
		return
	}

	aybusBot := bot.New(dg, userFollowChan, streamChangedChan, repository)
	aybusBot.Start()

	twitchWebhookManager := twitch.NewManager(twitchClientSecret, twitchClientId, userFollowChan, streamChangedChan, hubSecret, baseApiAddress)
	err = twitchWebhookManager.Start()

	log.Println("AYBUSH BOT is now running. Press CTRL + C to interrupt.")
	signalHandler := make (chan os.Signal)
	signal.Notify(signalHandler, os.Interrupt, os.Kill, syscall.SIGUSR1, syscall.SIGTERM)
	receivedSignal := <-signalHandler

	log.Printf("%v signal received. Gracefully shutting down the application.", receivedSignal)
	aybusBot.Stop()
	twitchWebhookManager.Stop()

	log.Printf("Application exited.")
}