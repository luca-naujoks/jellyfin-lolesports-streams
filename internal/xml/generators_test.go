package xml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"jellyfin-lolesports-streams/internal/types"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestGenerateTVGuide(t *testing.T) {
	startTime, _ := time.Parse("15:04:05", "15:15:15")
	applicationUrl := "https://jellybridge.de"
	channels := []types.Channel{
		{
			ID:       "riotgames",
			Name:     "Riot Games",
			Provider: "twitch",
			Leagues: []types.League{
				{
					Id:       "worlds",
					Slug:     "worlds",
					Name:     "worlds",
					Region:   "",
					ImageUrl: "",
				},
			},
			Language: "en",
			IconUrl:  "https://placeholder.com",
		},
	}
	programs := []types.Event{
		{
			ID:        "115660540726095109",
			League:    "worlds",
			Stage:     "Quarterfinals",
			StartTime: startTime,
			State:     "completed",
			Teams: []types.Team{
				{
					Name:  "Invictus Gaming",
					Code:  "IG",
					Image: "https://static.lolesports.com/teams/1634762917340_300px-Invictus_Gaming_logo.png",
				},
				{
					Name:  "TOP ESPORTS",
					Code:  "TES",
					Image: "https://static.lolesports.com/teams/1592592064571_TopEsportsTES-01-FullonDark.png",
				},
			},
			BestOf: 5,
		}, {
			ID:        "115660540726095109",
			League:    "worlds",
			Stage:     "Quarterfinals",
			StartTime: startTime,
			State:     "unstarted",
			Teams: []types.Team{
				{
					Name:  "Gen.G Esports",
					Code:  "Gen.G",
					Image: "https://static.lolesports.com/teams/1773829250929_GENGLOGO_GOLD.png",
				},
				{
					Name:  "LYON",
					Code:  "LYON",
					Image: "https://static.lolesports.com/teams/1743717443673_isotypelyon-03.png",
				},
			},
			BestOf: 5,
		},
	}

	expectedResult, err := os.ReadFile("../../testdata/tvguide.xml")
	if err != nil {
		t.Fatal(err)
	}

	res, err := GenerateTVGuide(channels, programs, applicationUrl)
	if err != nil {
		t.Fatal(err)
	}

	var expected, actual xmlTV

	err = xml.Unmarshal(expectedResult, &expected)
	if err != nil {
		t.Fatal(err)
	}

	err = xml.Unmarshal(res, &actual)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected and Generated XML do not Match")
	}
}

func TestGenerateM3u(t *testing.T) {
	baseUrl := "https://jellybridge.de/tv/stream/"
	channels := []types.Channel{
		{
			ID:       "riotgames",
			Name:     "Riot Games",
			Provider: "twitch",
			Leagues:  nil,
			Language: "en",
			IconUrl:  "https://placeholder.com",
		}, {
			ID:       "caedrel",
			Name:     "Caedrel",
			Provider: "twitch",
			Leagues:  nil,
			Language: "en",
			IconUrl:  "https://placeholder.com",
		},
	}

	expected, err := os.ReadFile("../../testdata/channels.m3u")
	if err != nil {
		t.Fatal(err)
	}

	result := GenerateM3u(channels, baseUrl)

	if !bytes.Equal(bytes.TrimSpace(expected), bytes.TrimSpace(result)) {
		fmt.Println(expected)
		fmt.Println(result)

		t.Errorf("Generated M3u channel List does not match the expected Result")
	}
}

func TestGenerateDescription(t *testing.T) {
	events := []types.Event{
		{
			ID:        "123",
			League:    "worlds",
			Stage:     "Swiss",
			StartTime: time.Now(),
			Teams: []types.Team{
				{
					Name: "Team 1",
				}, {
					Name: "Team 2",
				},
			},
			BestOf: 3,
		}, {
			ID:        "123",
			League:    "worlds",
			Stage:     "Quarterfinals",
			StartTime: time.Now(),
			Teams: []types.Team{
				{
					Name: "Team 1",
				}, {
					Name: "Team 2",
				},
			},
			BestOf: 5,
		}, {
			ID:        "123",
			League:    "lck",
			Stage:     "Summer Split",
			StartTime: time.Now(),
			Teams: []types.Team{
				{
					Name: "Team 1",
				}, {
					Name: "Team 2",
				},
			},
			BestOf: 1,
		},
	}
	expectedResults := []string{
		"Watch Team 1 face off Team 2 at worlds - Swiss in a Best-of-3 series",
		"Watch Team 1 face off Team 2 at worlds - Quarterfinals in a Best-of-5 series",
		"Watch Team 1 face off Team 2 at lck - Summer Split in a single-game match",
	}

	for i := range events {
		desc := generateDescription(events[i])
		if desc != expectedResults[i] {
			t.Errorf("%v did not match the expected Result: %s, expected: %s", i, desc, expectedResults[i])
		}
	}
}
