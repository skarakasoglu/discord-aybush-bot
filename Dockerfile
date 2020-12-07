FROM golang:1.15

RUN go get -u github.com/bwmarrin/discordgo && go get -u github.com/clinet/discordgo-embed && \
    go get -u github.com/disintegration/imaging && go get -u github.com/fogleman/gg && \
    go get -u github.com/gin-gonic/gin && go get -u github.com/spf13/viper && \
    go get -u mvdan.cc/xurls

WORKDIR /go/src/github.com/skarakasoglu/discord-aybush-bot/
ADD . /go/src/github.com/skarakasoglu/discord-aybush-bot/

# is going to be used with twitch webhooks.
EXPOSE 8080:8080

ENV applicationMode Debug
ENV applicationName "Discord Aybush Bot"
ENV discordAccessToken example_discord_access_token
ENV twitchAccessToken example_twitch_access_token
ENV twitchClientId example_twitch_client_id

CMD ["make", "all"]