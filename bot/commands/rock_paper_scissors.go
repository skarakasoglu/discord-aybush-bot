package commands

import (
	"bytes"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/skarakasoglu/discord-aybush-bot/configuration"
	"image"
	"image/jpeg"
	"log"
	"math"
	"math/rand"
	"strings"
	"time"
)

type gameElement string

const (
	rock gameElement = "rock"
	scissors gameElement = "scissors"
	paper gameElement = "paper"
)

type winner int

const (
	host winner = iota
	away
	draw
)
var (
	winnerMap = map[winner]string{
		host: "Host",
		away: "Away",
		draw: "Draw",
	}
)

func (w winner) String() string {
	str, _ := winnerMap[w]
	return str
}

type gameResult []gameElement

var (
	// Win conditions
	rockScissors  = gameResult{rock, scissors}
	rockPaper     = gameResult{paper, rock}
	paperScissors = gameResult{scissors, paper}

	// Draw conditions
	rockRock         = gameResult{rock, rock}
	scissorsScissors = gameResult{scissors, scissors}
	paperPaper       = gameResult{paper, paper}

	resultPossibilities = []winner{host, away, draw}

	winGame  = []gameResult{rockScissors, rockPaper, paperScissors}
	drawGame = []gameResult{rockRock, scissorsScissors, paperPaper}
)

type resultImageConfiguration struct{
	imageWidth int
	imageHeight int

	// General configurations
	avatarWidth int
	avatarHeight int
	elementWidth int
	elementHeight int
	arcRadius float64

	//Font configurations
	fontR float64
	fontG float64
	fontB float64
	fontSize float64
}

type resultImageParams struct{
	BackgroundImagePath string
	Configuration resultImageConfiguration
	HostParams playerImageParams
	AwayParams playerImageParams
}

type playerImageParams struct{
	element string
	username string
	avatar image.Image
	avatarX int
	avatarY int
	elementX int
	elementY int
	usernameX float64
	usernameY float64
	arcX float64
	arcY float64
}

type RockPaperScissorsCommand struct {
	session *discordgo.Session
	randSource rand.Source
	rnd *rand.Rand
}

func NewRockPaperScissorsCommand(session *discordgo.Session) Command{
	return &RockPaperScissorsCommand{
		session: session,
		randSource: rand.NewSource(time.Now().UnixNano()),
		rnd: rand.New(randSource),
	}
}

func (cmd *RockPaperScissorsCommand) Name() string {
	return "taş-kağıt-makas"
}

func (cmd *RockPaperScissorsCommand) Execute(message *discordgo.Message) (string, error) {
	arguments := strings.Split(message.Content, " ")[1:]

	if len(arguments) < 1 || len(message.Mentions) != 1 {
		return fmt.Sprintf("<@%v>, %v", message.Author.ID, cmd.Usage()), nil
	}

	var chosenGameResult gameResult

	hostPlayer := message.Author
	awayPlayer := message.Mentions[0]

	chosenWinnerIndex := rnd.Intn(len(resultPossibilities))

	backgroundImage := ""
	hostUsername := fmt.Sprintf("%v#%v", hostPlayer.Username, hostPlayer.Discriminator)
	awayUsername := fmt.Sprintf("%v#%v", awayPlayer.Username, awayPlayer.Discriminator)
	hostAvatar, err := cmd.session.UserAvatar(hostPlayer.ID)
	if err != nil {
		log.Printf("Error on obtaining host player avatar: %v", err)
	}

	awayAvatar, err := cmd.session.UserAvatar(awayPlayer.ID)
	if err != nil {
		log.Printf("Error on obtaining away player avatar: %v", err)
	}

	var winner winner

	if resultPossibilities[chosenWinnerIndex] == draw {
		chosenGamePossibilityIndex := rnd.Intn(len(drawGame))
		chosenGameResult = drawGame[chosenGamePossibilityIndex]
		winner = draw
		backgroundImage = fmt.Sprintf("%v/%v", configuration.Manager.BaseImagePath, configuration.Manager.RockPaperScissors.Draw)
	} else {
		chosenGamePossibilityIndex := rnd.Intn(len(winGame))
		chosenGameResult = winGame[chosenGamePossibilityIndex]

		if resultPossibilities[chosenWinnerIndex] == host {
			winner = host
			backgroundImage = fmt.Sprintf("%v/%v", configuration.Manager.BaseImagePath, configuration.Manager.RockPaperScissors.HostWins)
		} else {
			winner = away
			backgroundImage = fmt.Sprintf("%v/%v", configuration.Manager.BaseImagePath, configuration.Manager.RockPaperScissors.AwayWins)
			chosenGameResult = gameResult{chosenGameResult[away], chosenGameResult[host]}
		}
	}

	log.Printf("Rock-Paper-Scissors was played. Host: %v#%v %v, Away: %v#%v %v, Winner: %v",
		hostPlayer.Username, hostPlayer.Discriminator, chosenGameResult[host],
		awayPlayer.Username, awayPlayer.Discriminator, chosenGameResult[away],
		winner.String())

	hostElementPath := fmt.Sprintf("%v/%v.png", configuration.Manager.BaseImagePath, chosenGameResult[host])
	awayElementPath := fmt.Sprintf("%v/%v.png", configuration.Manager.BaseImagePath, chosenGameResult[away])

	params := resultImageParams{
		BackgroundImagePath:      backgroundImage,
		Configuration:            resultImageConfiguration{
			imageWidth:    800,
			imageHeight:   250,
			avatarWidth:   140,
			avatarHeight:  140,
			elementWidth:  180,
			elementHeight: 180,
			arcRadius:     70,
			fontR:         0,
			fontG:         0,
			fontB:         0,
			fontSize:      25,
		},
		HostParams:               playerImageParams{
			element: hostElementPath,
			username: hostUsername,
			avatar: hostAvatar,
			avatarX: 40,
			avatarY: 40,
			elementX: 210,
			elementY: 30,
			usernameX: 20,
			usernameY: 220,
			arcX: 110,
			arcY: 110,
		},
		AwayParams:               playerImageParams{
			element: awayElementPath,
			username: awayUsername,
			avatar: awayAvatar,
			avatarX: 620,
			avatarY: 40,
			elementX: 430,
			elementY: 30,
			usernameX: 600,
			usernameY: 220,
			arcX: 690,
			arcY: 110,
		},
	}


	resultImg, err := cmd.createResultImage(params)
	if err != nil {
		log.Printf("Error on creating image: %v", err)
		return "", err
	}

	imageBuffer := new(bytes.Buffer)
	err = jpeg.Encode(imageBuffer, resultImg, nil)
	if err != nil {
		log.Printf("Error on writing image to buffer: %v", err)
		return "", err
	}

	_, err = cmd.session.ChannelFileSend(message.ChannelID, "oyun_sonucu.png", imageBuffer)
	if err != nil {
		log.Printf("Error on sending file to channel: %v", err)
		return "", err
	}

	return "", nil
}

func (cmd *RockPaperScissorsCommand) Usage() string {
	usageType := fmt.Sprintf("!%v <kullanıcı-adı>", cmd.Name())
	return fmt.Sprintf("**bu komutu**\n%v şeklinde kullanabilirsiniz.", usageType)
}

func (cmd *RockPaperScissorsCommand) createResultImage(params resultImageParams) (image.Image, error) {

	imageContext := gg.NewContext(params.Configuration.imageWidth, params.Configuration.imageHeight)

	backgroundImage, err := gg.LoadImage(params.BackgroundImagePath)
	if err != nil {
		log.Printf("Error on loading image: %v" ,err)
		return nil, err
	}

	backgroundImageResized := imaging.Resize(backgroundImage, params.Configuration.imageWidth, params.Configuration.imageHeight, imaging.Lanczos)
	imageContext.DrawImage(backgroundImageResized, 0, 0)

	hostElement, err := gg.LoadImage(params.HostParams.element)
	if err != nil {
		log.Printf("Error on loading host element image: %v", err)
		return nil, err
	}

	hostElementResized := imaging.Resize(hostElement, params.Configuration.elementWidth, params.Configuration.elementHeight, imaging.Lanczos)
	imageContext.DrawImage(hostElementResized, params.HostParams.elementX, params.HostParams.elementY)

	awayElement, err := gg.LoadImage(params.AwayParams.element)
	if err != nil {
		log.Printf("Error on loading away element image: %v", err)
		return nil, err
	}

	awayElementResized := imaging.Resize(awayElement, params.Configuration.elementWidth, params.Configuration.elementHeight, imaging.Lanczos)
	imageContext.DrawImage(awayElementResized, params.AwayParams.elementX, params.AwayParams.elementY)


	imageContext.SetRGB(params.Configuration.fontR, params.Configuration.fontG, params.Configuration.fontB)
	err = imageContext.LoadFontFace("fonts/Roboto-Medium.ttf", params.Configuration.fontSize)
	if err != nil {
		log.Printf("Error on loading font: %v", err)
	}

	imageContext.DrawString(params.HostParams.username, params.HostParams.usernameX, params.HostParams.usernameY)
	imageContext.DrawString(params.AwayParams.username, params.AwayParams.usernameX, params.AwayParams.usernameY)

	imageContext.NewSubPath()
	imageContext.DrawArc(params.HostParams.arcX, params.HostParams.arcY, params.Configuration.arcRadius, 0, math.Pi * 2)
	imageContext.DrawArc(params.AwayParams.arcX, params.AwayParams.arcY, params.Configuration.arcRadius, 0, math.Pi * 2)
	imageContext.ClosePath()
	imageContext.Clip()

	hostAvatarResized := imaging.Resize(params.HostParams.avatar, params.Configuration.avatarWidth, params.Configuration.avatarHeight, imaging.Lanczos)
	imageContext.DrawImage(hostAvatarResized, params.HostParams.avatarX, params.HostParams.avatarY)

	awayAvatarResized := imaging.Resize(params.AwayParams.avatar, params.Configuration.avatarWidth, params.Configuration.avatarHeight, imaging.Lanczos)
	imageContext.DrawImage(awayAvatarResized, params.AwayParams.avatarX, params.AwayParams.avatarY)

	return imageContext.Image(), nil
}