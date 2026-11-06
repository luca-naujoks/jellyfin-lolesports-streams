package main

import (
	"embed"
	"fmt"
	"jellyfin-lolesports-streams/internal/api"
	"jellyfin-lolesports-streams/internal/store"
	"jellyfin-lolesports-streams/internal/types"
	"net/http"
	"os"
	"time"
)

//go:embed assets
var assets embed.FS

//go:embed web-app/dist/client
var webInterface embed.FS

func main() {
	fmt.Println("Loading Configuration from environment variables:")
	fmt.Println("-------------------------------------------------")
	config := loadConfiguration()
	fmt.Println("-------------------------------------------------")

	serverErrors := make(chan error, 1)
	httpClient := http.DefaultClient

	fmt.Print("Creating New SQliteStore...")
	SQLStore, err := store.NewSqliteStore(config.DevMode)
	if err != nil {
		panic(err)
	}
	fmt.Print("\r\033[2K")
	fmt.Printf("Creating New SQliteStore... OK\n")

	fmt.Print("Creating New Handler...")
	handler := api.NewHandler(httpClient, config, SQLStore, assets)
	fmt.Print("\r\033[2K")
	fmt.Printf("Creating New Handler... OK\n")

	fmt.Print("Creating New Web Router...")
	srv := api.NewRouter(handler, webInterface)
	fmt.Print("\r\033[2K")
	fmt.Printf("Creating New Web Router... OK\n")

	fmt.Print("Starting Web Server...")
	go srv.Run(serverErrors)
	fmt.Print("\r\033[2K")
	fmt.Printf("Starting Web Server... OK\n")

	fmt.Println("Application startup Complete")

	srv.ShutdownHttpServer()
}

func loadConfiguration() types.MainConfig {
	var config types.MainConfig

	defaults := types.MainConfig{
		ApplicationUrl: "http://localhost:80",
		CacheTTL:       15 * time.Minute,

		TwitchClientId:   "kimne78kx3ncx6brgo4mv6wki5h1ko",
		TwitchUserAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:157.0) Gecko/20100101 Firefox/157.0",
		LolEsportsSecret: "0TvQnueqKa5mxJntVWt0w4LpLfEkrV1Ta8rQBb9Z",
		LolEsportsApiUrl: "https://esports-api.lolesports.com/persisted/gw",
	}

	config.DevMode = getEnv("DevMode", "false") == "true"

	config.ApplicationUrl = getEnv("ApplicationUrl", defaults.ApplicationUrl)
	config.CacheTTL = defaults.CacheTTL

	config.TwitchClientId = getEnv("TwitchClientId", defaults.TwitchClientId)
	config.TwitchUserAgent = getEnv("DefaultUserAgent", defaults.TwitchUserAgent)

	config.LolEsportsSecret = getEnv("LolEsportsSecret", defaults.LolEsportsSecret)
	config.LolEsportsApiUrl = getEnv("LolEsportsApiUrl", defaults.LolEsportsApiUrl)

	return config
}

func getEnv(envName string, defaultValue string) string {
	envValue := os.Getenv(envName)
	if envValue == "" {
		message := fmt.Sprintf("%s: %s (Default Value)", envName, defaultValue)
		fmt.Println(message)
		return defaultValue
	}

	message := fmt.Sprintf("%s: %s", envName, envValue)
	fmt.Println(message)
	return envValue
}
