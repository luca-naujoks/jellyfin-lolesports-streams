package lolesports

import (
	"context"
	"net/http"
	"time"
)

type Client struct {
	http *http.Client
	ctx  context.Context
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
