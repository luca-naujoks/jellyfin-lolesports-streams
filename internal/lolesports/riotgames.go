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

var (
	baseUrl = "https://esports-api.lolesports.com/persisted/gw"

	apiKeyHeader = "x-api-key"
	secret       = "0TvQnueqKa5mxJntVWt0w4LpLfEkrV1Ta8rQBb9Z"
)

func NewRiotClient(client *http.Client) *Client {
	if client == nil {
		client = http.DefaultClient
	}
	return &Client{http: client, ctx: context.Background()}
}

func (c Client) httpRequest(req *http.Request) ([]byte, error) {
	req.Header.Set(apiKeyHeader, secret)

	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

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
	req, err := http.NewRequestWithContext(c.ctx, http.MethodGet, fmt.Sprintf("%s/getLeagues?hl=en-US", baseUrl), nil)
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
	req, err := http.NewRequestWithContext(c.ctx, http.MethodGet, fmt.Sprintf("%s/getSchedule?hl=en-US&leagueId=%s", baseUrl, ids), nil)
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

	// Filter Evets that happened before a week
	var events []types.Event

	for i := range response.Data.Schedule.Events {
		//if the Event is older than a week ago we ignore it
		if response.Data.Schedule.Events[i].StartTime.Before(time.Now().AddDate(0, 0, -7)) || response.Data.Schedule.Events[i].Type != "match" {
			continue
		}

		// Every Event here is in range of a week old or in the future
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
