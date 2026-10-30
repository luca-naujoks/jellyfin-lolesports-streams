package types

import (
	"time"
)

type Event struct {
	ID        string
	League    string
	Stage     string
	StartTime time.Time
	State     string
	Teams     []Team
	BestOf    int
}

type Team struct {
	Name  string `json:"name"`
	Code  string `json:"code"`
	Image string `json:"image"`
}

type Channel struct {
	ID       string // Required
	Name     string // Required
	Provider string
	Leagues  []League
	Language string
	IconUrl  string
}

type League struct {
	Id       string `json:"id"`
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Region   string `json:"region"`
	ImageUrl string `json:"image"`
	Priority int    `json:"priority"`
}

type BroadCastChannel struct {
	ID      string
	Name    string
	Leagues []string
}
