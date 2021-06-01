package level

import (
	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"image"
	"log"
	"math"
)

type RankImageOptions struct{
	BackgroundImagePath string
	FontFace string

	Width int
	Height int
	AvatarWidth int
	AvatarHeight int
	AvatarArcX float64
	AvatarArcY float64
	AvatarRadius float64
	AvatarX int
	AvatarY int

	RankTextOptions ShadowedTextOptions
	UsernameTextOptions ShadowedTextOptions
	RoleTextOptions ShadowedTextOptions
	LevelTextOptions ShadowedTextOptions
	ExpBarOptions
	CurrentExpOptions ShadowedTextOptions
	CurrentLevelOptions ShadowedTextOptions
	NextLevelOptions ShadowedTextOptions

	Avatar image.Image
}

type ShadowedTextOptions struct {
	Text string
	StrokeSize int
	X float64
	Y float64
	Ax float64
	Ay float64
	ShadowOptions ColorOptions
	TextOptions ColorOptions
}

type ColorOptions struct {
	R int
	G int
	B int
	Alpha int
}

type ExpBarOptions struct{
	X float64
	Y float64
	Width float64
	Height float64
	Radius float64
	StrokeSize int
	CurrentExperience uint64
	CurrentLevelRequiredExperience uint64
	RequiredExperience uint64
	ShadowOptions ColorOptions
	EmptyBarOptions ColorOptions
	FilledBarOptions ColorOptions
}

func createExpBar(dc *gg.Context, options ExpBarOptions) {
	dc.SetRGBA255(options.ShadowOptions.R, options.ShadowOptions.G, options.ShadowOptions.B, options.ShadowOptions.Alpha)
	n := options.StrokeSize
	for dy := -n; dy <= n; dy++ {
		for dx := -n; dx <=n; dx++ {
			if dx*dx+dy*dy >= n*n {
				continue
			}

			x := options.X + float64(dx)
			y := options.Y + float64(dy)
			dc.DrawRoundedRectangle(x, y, options.Width, options.Height, options.Radius)
			dc.Fill()
		}
	}

	dc.DrawRoundedRectangle(options.X, options.Y, options.Width, options.Height, options.Radius)
	dc.SetRGBA255(options.EmptyBarOptions.R,options.EmptyBarOptions.G,options.EmptyBarOptions.B, options.EmptyBarOptions.Alpha)
	dc.Fill()

	width := (float64((options.CurrentExperience - options.CurrentLevelRequiredExperience) * 100) / float64(options.RequiredExperience - options.CurrentLevelRequiredExperience)) * (options.Width / 100)

	dc.DrawRoundedRectangle(options.X, options.Y, width, options.Height, options.Radius)
	dc.SetRGBA255(options.FilledBarOptions.R,options.FilledBarOptions.G,options.FilledBarOptions.B, options.FilledBarOptions.Alpha)
	dc.Fill()
}

func createShadowedText(dc *gg.Context, options ShadowedTextOptions) {
	dc.SetRGBA255(options.ShadowOptions.R, options.ShadowOptions.G,  options.ShadowOptions.B, options.ShadowOptions.Alpha)
	s := options.Text
	n := options.StrokeSize
	for dy := -n; dy <= n; dy++ {
		for dx := -n; dx <=n; dx++ {
			if dx*dx+dy*dy >= n*n {
				continue
			}

			strokeX := options.X + float64(dx)
			strokeY := options.Y + float64(dy)
			dc.DrawStringAnchored(s, strokeX, strokeY, options.Ax, options.Ay)
		}
	}

	dc.SetRGBA255(options.TextOptions.R, options.TextOptions.G, options.TextOptions.B, options.TextOptions.Alpha)
	dc.DrawStringAnchored(s, options.X, options.Y, options.Ax, options.Ay)
}

func createRankImage(options RankImageOptions) image.Image{
	imageContext := gg.NewContext(options.Width, options.Height)

	backgroundImage, err := gg.LoadImage(options.BackgroundImagePath)
	if err != nil {
		log.Printf("Error on loading background image: %v", err)
		return imageContext.Image()
	}

	imageContext.DrawImage(backgroundImage, 0, 0)

	err = imageContext.LoadFontFace(options.FontFace, 25)
	createShadowedText(imageContext, options.RankTextOptions)
	createShadowedText(imageContext, options.UsernameTextOptions)

	err = imageContext.LoadFontFace(options.FontFace, 18)
	createShadowedText(imageContext, options.RoleTextOptions)
	err = imageContext.LoadFontFace(options.FontFace, 25)

	createShadowedText(imageContext, options.LevelTextOptions)
	createExpBar(imageContext, options.ExpBarOptions)

	err = imageContext.LoadFontFace(options.FontFace, 14)
	createShadowedText(imageContext, options.CurrentExpOptions)

	err = imageContext.LoadFontFace(options.FontFace, 25)
	createShadowedText(imageContext, options.CurrentLevelOptions)
	createShadowedText(imageContext, options.NextLevelOptions)

	imageContext.NewSubPath()
	imageContext.DrawArc(options.AvatarArcX, options.AvatarArcY, options.AvatarRadius, 0, math.Pi * 2)
	imageContext.ClosePath()
	imageContext.Clip()

	avatarResized := imaging.Resize(options.Avatar, options.AvatarWidth, options.AvatarHeight, imaging.Lanczos)
	imageContext.DrawImage(avatarResized, options.AvatarX, options.AvatarY)

	return imageContext.Image()
}