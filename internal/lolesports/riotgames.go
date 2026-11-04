package lolesports

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"jellyfin-lolesports-streams/internal/types"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	http   *http.Client
	config Config
}

type Config struct {
	BaseUrl string
	Secret  string
}

type ScheduleResponse struct {
	Data struct {
		Schedule struct {
			Pages struct {
				Older string `json:"older"`
				Newer any    `json:"newer"`
			} `json:"pages"`
			Events []struct {
				StartTime time.Time `json:"startTime"`
				State     string    `json:"state"`
				Type      string    `json:"type"`
				BlockName string    `json:"blockName"`
				League    struct {
					Name string `json:"name"`
					Slug string `json:"slug"`
				} `json:"league"`
				Match struct {
					ID    string   `json:"id"`
					Flags []string `json:"flags"`
					Teams []struct {
						Name   string `json:"name"`
						Code   string `json:"code"`
						Image  string `json:"image"`
						Result struct {
							Outcome  string `json:"outcome"`
							GameWins int    `json:"gameWins"`
						} `json:"result"`
						Record struct {
							Wins   int `json:"wins"`
							Losses int `json:"losses"`
						} `json:"record"`
					} `json:"teams"`
					Strategy struct {
						Type  string `json:"type"`
						Count int    `json:"count"`
					} `json:"strategy"`
				} `json:"match"`
			} `json:"events"`
		} `json:"schedule"`
	} `json:"data"`
}

func NewRiotClient(client *http.Client, config Config) *Client {
	if client == nil {
		client = http.DefaultClient
	}
	return &Client{http: client, config: config}
}

func (c Client) httpRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("x-api-key", c.config.Secret)

	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Error closing response Body: %s", err.Error())
		}
	}(res.Body)

	if res.StatusCode >= 400 {
		message := fmt.Sprintf("error Code %v received", res.StatusCode)
		return nil, errors.New(message)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// RequestAllLeagues Request Leagues eg. LCK Worlds, MSI, LEC...
func (c Client) RequestAllLeagues() ([]types.League, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf("%s/getLeagues?hl=en-US", c.config.BaseUrl), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.httpRequest(req)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data struct {
			Leagues []types.League `json:"leagues"`
		} `json:"data"`
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return response.Data.Leagues, nil
}

func (c Client) RequestLeagueSchedule(leagueIds []string) ([]types.Event, error) {
	ids := strings.Join(leagueIds, ",")
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf("%s/getSchedule?hl=en-US&leagueId=%s", c.config.BaseUrl, ids), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.httpRequest(req)
	if err != nil {
		return nil, err
	}

	var response ScheduleResponse

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	var events []types.Event

	for i := range response.Data.Schedule.Events {
		if response.Data.Schedule.Events[i].Type != "match" {
			continue
		}

		if response.Data.Schedule.Events[i].StartTime.Before(time.Now().AddDate(0, 0, -1)) {
			continue
		}

		event := response.Data.Schedule.Events[i]

		eventStruct := types.Event{
			ID:        event.Match.ID,
			League:    event.League.Slug,
			Stage:     event.BlockName,
			StartTime: event.StartTime,
			State:     event.State,
			Teams: []types.Team{
				{
					Name:  event.Match.Teams[0].Name,
					Code:  event.Match.Teams[0].Code,
					Image: event.Match.Teams[0].Image,
				}, {
					Name:  event.Match.Teams[1].Name,
					Code:  event.Match.Teams[1].Code,
					Image: event.Match.Teams[1].Image,
				},
			},
			BestOf: event.Match.Strategy.Count,
		}

		events = append(events, eventStruct)
	}

	return events, nil
}
