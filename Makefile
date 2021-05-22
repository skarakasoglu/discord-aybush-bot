build:
	echo "Building ${applicationName}..."
	go build -o ${executablePath}/${executableName} main.go

run:
	echo "Running ${applicationName} in ${applicationMode} mode..."
	${executablePath}/${executableName} --discord-token ${discordAccessToken} \
		--twitch-client-secret ${twitchClientSecret} --twitch-client-id ${twitchClientId} --twitch-refresh-token ${twitchRefreshToken} \
		--hub-secret ${webhookHubSecret} --base-api-address ${baseApiAddress} --db-ip-address ${dbIpAddress} --db-port ${dbPort} --db-username ${dbUsername} --db-password ${dbPassword} --db-name ${dbName}

all: build run