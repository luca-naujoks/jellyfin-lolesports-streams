package lolesports

import (
	"image"
	"math"
	"os"

	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

func calculateFontYCenter(baseImage *image.RGBA, face font.Face) int {
	imageHeight := baseImage.Bounds().Dy()
	metrics := face.Metrics()

	baselineOffset := metrics.Ascent - metrics.Descent

	totalCenter := baselineOffset/2 + fixed.I(imageHeight)/2

	return totalCenter.Ceil()
}

func calculateScaledDimensions(originalWidth int, originalHeight int, maxWidth int, maxHeight int) (int, int) {
	widthScale := float64(maxWidth) / float64(originalWidth)
	heightScale := float64(maxHeight) / float64(originalHeight)

	scale := min(widthScale, heightScale)

	return int(math.Round(float64(originalWidth) * scale)), int(math.Round(float64(originalHeight) * scale))
}

func getFontFace(fontPath string) (font.Face, error) {
	file, err := os.ReadFile(fontPath)
	if err != nil {
		return nil, err
	}

	parse, err := opentype.Parse(file)
	if err != nil {
		return nil, err
	}

	face, err := opentype.NewFace(parse, &opentype.FaceOptions{
		Size:    32,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, err
	}

	return face, nil
}

func resizeImage(src image.Image, maxWidth, maxHeight int) *image.RGBA {
	width, height := calculateScaledDimensions(src.Bounds().Dx(), src.Bounds().Dy(), maxWidth, maxHeight)

	dstRect := image.Rect(0, 0, width, height)
	scaledImage := image.NewRGBA(dstRect)

	draw.BiLinear.Scale(scaledImage, dstRect, src, src.Bounds(), draw.Over, nil)

	return scaledImage
}
