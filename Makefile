build:
	echo "Building ${applicationName}..."
	go build -o ${executablePath}/${executableName} main.go

run:
	echo "Running ${applicationName} in ${applicationMode} mode..."
	${executablePath}/${executableName} --discord-token ${discordAccessToken} \
		--twitch-client-secret ${twitchClientSecret} --twitch-client-id ${twitchClientId} --twitch-refresh-token ${twitchRefreshToken} \
		--hub-secret ${webhookHubSecret} --db-ip-address ${dbIpAddress} --db-port ${dbPort} --db-username ${dbUsername} --db-password ${dbPassword} --db-name ${dbName} \
		--cert-file ${certFile} --key-file ${keyFile} --streamlabs-access-token ${streamlabsAccessToken} --shopier-username ${shopierUsername} --shopier-key ${shopierKey}

all: build run