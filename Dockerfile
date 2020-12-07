FROM golang:1.15

WORKDIR /go/src/github.com/skarakasoglu/discord-aybush-bot/
ADD . /go/src/github.com/skarakasoglu/discord-aybush-bot/

RUN go get -u github.com/fogleman/gg && go get -u github.com/disintegration/imaging && \\
    go get -u github.com/bwmarrin/discordgo

# is going to be used with twitch webhooks.
EXPOSE 8080:8080

CMD ["go", "run", "main.go"]