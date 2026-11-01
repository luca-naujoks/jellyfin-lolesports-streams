package lolesports

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"jellyfin-lolesports-streams/internal/types"
	"net/http"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// GenerateVsImage Generates the Program Icon
func GenerateVsImage(teams []types.Team) ([]byte, error) {
	var imageBuffer bytes.Buffer
	var logos []image.Image

	width := 400
	height := 200
	baseImage := image.NewRGBA(image.Rect(0, 0, width, height))

	uniform := image.NewUniform(color.RGBA{
		R: 18,
		G: 24,
		B: 38,
		A: 255,
	})

	draw.Draw(baseImage, baseImage.Bounds(), uniform, image.Point{}, draw.Src)

	for i := range teams {
		logo, err := requestImage(teams[i].Image)
		if err != nil {
			return nil, err
		}

		logos = append(logos, logo)
	}

	face, err := getFontFace("./assets/Roboto-BoldItalic.ttf")
	if err != nil {
		return nil, err
	}
	defer func(face font.Face) {
		err := face.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
	}(face)

	drawOverlay(baseImage, []image.Image{logos[0], logos[1]})

	yCenter := calculateFontYCenter(baseImage, face)

	addText(baseImage, "VS", yCenter, face)
	addText(baseImage, fmt.Sprintf("%s vs %s", teams[0].Code, teams[1].Code), 180, face)

	err = png.Encode(&imageBuffer, baseImage)

	return imageBuffer.Bytes(), err
}

func requestImage(imageUrl string) (image.Image, error) {
	client := http.DefaultClient

	res, err := client.Get(imageUrl)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
	}(res.Body)

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("image request failed with status: %v", res.StatusCode)
	}

	return png.Decode(res.Body)
}

func drawOverlay(baseImage *image.RGBA, logos []image.Image) {
	centerClearance := 60
	biX := baseImage.Bounds().Dx()
	biY := baseImage.Bounds().Dy()

	maxWidth := 100
	maxHeight := 100

	// Logo 1
	resizedLogo := resizeImage(logos[0], maxWidth, maxHeight)
	startX := biX/2 - centerClearance - resizedLogo.Bounds().Dx()
	startY := (biY - resizedLogo.Bounds().Dy()) / 2
	endX := startX + resizedLogo.Bounds().Dx()
	endY := startY + resizedLogo.Bounds().Dy()

	rec := image.Rect(startX, startY, endX, endY)

	draw.Draw(baseImage, rec, resizedLogo, resizedLogo.Bounds().Min, draw.Over)

	// Logo 2
	resizedLogo = resizeImage(logos[1], maxWidth, maxHeight)

	startX = biX/2 + centerClearance
	startY = (biY - resizedLogo.Bounds().Dy()) / 2
	endX = startX + resizedLogo.Bounds().Dx()
	endY = startY + resizedLogo.Bounds().Dy()

	rec = image.Rect(startX, startY, endX, endY)

	draw.Draw(baseImage, rec, resizedLogo, resizedLogo.Bounds().Min, draw.Over)

}

func addText(baseImage *image.RGBA, text string, y int, face font.Face) {
	col := color.White

	drawer := font.Drawer{
		Dst:  baseImage,
		Src:  image.NewUniform(col),
		Face: face,
	}

	textWidth := drawer.MeasureString(text)

	startX := (baseImage.Bounds().Dx() - textWidth.Ceil()) / 2

	drawer.Dot = fixed.Point26_6{
		X: fixed.I(startX),
		Y: fixed.I(y),
	}

	drawer.DrawString(text)
}
