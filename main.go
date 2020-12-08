package main

import (
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/skarakasoglu/discord-aybush-bot/bot"
	"github.com/skarakasoglu/discord-aybush-bot/configuration"
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
	twitchAccessToken string
	twitchClientId string
)

func init() {
	flag.StringVar(&discordAccessToken,"discord-token", "", "discord api application access token")
	flag.StringVar(&configurationFileName,"cfg-file", "config", "application configuration file name")
	flag.StringVar(&configurationFilePath, "cfg-file-path", ".", "application configuration file path")
	flag.StringVar(&twitchAccessToken, "twitch-token", "", "twitch api oauth token")
	flag.StringVar(&twitchClientId, "twitch-client-id", "", "twitch api client id")
	flag.Parse()

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

	aybusBot := bot.New(dg, userFollowChan, streamChangedChan)
	aybusBot.Start()

	twitchWebhookManager := twitch.NewManager(twitchAccessToken, twitchClientId, userFollowChan, streamChangedChan)
	err = twitchWebhookManager.Start()

	log.Println("AYBUŞ BOT is now running. Press CTRL + C to interrupt.")
	signalHandler := make (chan os.Signal)
	signal.Notify(signalHandler, os.Interrupt, os.Kill, syscall.SIGUSR1, syscall.SIGTERM)
	receivedSignal := <-signalHandler

	log.Printf("%v signal received. Gracefully shutting down the application.", receivedSignal)
	aybusBot.Stop()
	twitchWebhookManager.Stop()

	log.Printf("Application exited.")
}