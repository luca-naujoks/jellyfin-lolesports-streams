package lolesports

import (
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
