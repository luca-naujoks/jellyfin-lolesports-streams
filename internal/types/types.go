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
	ID       string   `json:"id"`   // Required
	Name     string   `json:"name"` // Required
	Provider string   `json:"provider"`
	Leagues  []League `json:"leagues"`
	Language string   `json:"language"`
	IconUrl  string   `json:"icon_url"`
}

type League struct {
	Id       string `json:"id"`
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Region   string `json:"region"`
	ImageUrl string `json:"image"`
}

type BroadCastChannel struct {
	ID      string
	Name    string
	Leagues []string
}
