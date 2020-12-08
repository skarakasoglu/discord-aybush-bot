build:
	echo "Building ${applicationName}..."
	go build -o ${executablePath}/${executableName} main.go

run:
	echo "Running ${applicationName} in ${applicationMode} mode..."
	bin/DiscordAybushBot --discord-token ${discordAccessToken} --twitch-token ${twitchAccessToken} --twitch-client-id ${twitchClientId}

all: build run