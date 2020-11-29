package main

import (
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/skarakasoglu/discord-aybush-bot/bot"
	"github.com/skarakasoglu/discord-aybush-bot/configuration"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	configurationFileName string
	configurationFilePath string
	accessToken string
)

func init() {
	flag.StringVar(&accessToken,"token", "", "discord api application access token")
	flag.StringVar(&configurationFileName,"cfg-file", "config", "application configuration file name")
	flag.StringVar(&configurationFilePath, "cfg-file-path", ".", "application configuration file path")
	flag.Parse()

	configuration.ReadConfigurationFile(configurationFilePath, configurationFileName)
}

func main() {
	dg, err := discordgo.New(fmt.Sprintf("Bot %v", accessToken))
	if err != nil {
		log.Fatalf("Failed to create: %v", err)
	}

	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)
	err = dg.Open()
	if err != nil {
		log.Fatalf("Failed to open websocket connection with Discord API: %v", err)
	}

	aybusBot := bot.New(dg)
	aybusBot.Start()


	log.Println("AYBUŞ BOT is now running. Press CTRL + C to interrupt.")
	signalHandler := make (chan os.Signal)
	signal.Notify(signalHandler, os.Interrupt, os.Kill, syscall.SIGSEGV, syscall.SIGHUP)
	receivedSignal := <-signalHandler

	log.Printf("%v signal received. Gracefully shutting down the application.", receivedSignal)
	aybusBot.Stop()
}