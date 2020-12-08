FROM golang:1.15

RUN DEBIAN_FRONTEND=noninteractive apt-get install -y tzdata
ENV TZ Asia/Istanbul

#Install go dependencies.
RUN go get -u github.com/bwmarrin/discordgo && go get -u github.com/clinet/discordgo-embed && \
    go get -u github.com/disintegration/imaging && go get -u github.com/fogleman/gg && \
    go get -u github.com/gin-gonic/gin && go get -u github.com/spf13/viper && \
    go get -u mvdan.cc/xurls

#Create group and user named aybush, do not create home directory, do not assign password to user.
RUN groupadd aybush && useradd -m -g aybush aybush
#Log in as aybush.
USER aybush

WORKDIR /go/src/github.com/skarakasoglu/discord-aybush-bot/
#Change owner and group of files as aybush.
ADD --chown=aybush:aybush . /go/src/github.com/skarakasoglu/discord-aybush-bot/

#Give the owner to execute the run.sh shell script.
RUN chmod 744 run.sh
RUN chown aybush:aybush /go/src/github.com/skarakasoglu/discord-aybush-bot/ && chmod 744 run.sh

# is going to be used with twitch webhooks.
EXPOSE 8080:8080

ENV applicationMode Debug
ENV applicationName "Discord Aybush Bot"
ENV executablePath bin
ENV executableName DiscordAybushBot
ENV discordAccessToken example_discord_access_token
ENV twitchAccessToken example_twitch_access_token
ENV twitchClientId example_twitch_client_id

CMD ["/bin/bash", "-c", "./run.sh"]