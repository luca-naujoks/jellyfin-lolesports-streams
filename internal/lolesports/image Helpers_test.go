package lolesports

import (
	"image"
	"testing"
)

type aspects struct {
	width          int
	height         int
	expectedWidth  int
	expectedHeight int
}

func TestAspectRatio(t *testing.T) {
	testCases := []aspects{
		{width: 500, height: 500, expectedHeight: 100, expectedWidth: 100},
		{width: 400, height: 200, expectedWidth: 100, expectedHeight: 50},
		{width: 200, height: 400, expectedWidth: 50, expectedHeight: 100},
		{width: 800, height: 400, expectedWidth: 100, expectedHeight: 50},
	}

	for i, testCase := range testCases {
		width, height := calculateScaledDimensions(testCase.width, testCase.height, 100, 100)
		if width != testCase.expectedWidth || height != testCase.expectedHeight {
			t.Errorf("testCase %v returned not the expected ratio. returned: %v, %v. Expected %v, %v", i, width, height, testCase.expectedWidth, testCase.expectedHeight)
		}
	}

}

func TestResizeImage(t *testing.T) {
	maxWidth := 100
	maxHeight := 100

	expectedX := 50
	expectedY := 100

	baseImage := image.NewRGBA(image.Rect(0, 0, 150, 300))

	rgba := resizeImage(baseImage, maxWidth, maxHeight)

	if rgba.Bounds().Dy() != expectedY || rgba.Bounds().Dx() != expectedX {
		t.Errorf("Expected: x:%v, y:%v\n Received: x:%v, y:%v", expectedX, expectedY, rgba.Bounds().Dx(), rgba.Bounds().Dy())
	}
}
