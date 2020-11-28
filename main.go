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
)

func init() {
	configurationFileName = *flag.String("cfg-file", "config", "application configuration file name")
	configurationFilePath = *flag.String("cfg-file-path", ".", "application configuration file path")
	flag.Parse()

	configuration.ReadConfigurationFile(configurationFilePath, configurationFileName)
}

func main() {

	dg, err := discordgo.New(fmt.Sprintf("Bot %v", configuration.Manager.Credentials.GetToken()))
	if err != nil {
		log.Fatalf("Failed to create: %v", err)
	}

	err = dg.Open()
	if err != nil {
		log.Fatalf("Failed to open websocket connection: %v", err)
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