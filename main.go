package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"jellyfin-lolesports-streams/internal/api"
	"jellyfin-lolesports-streams/internal/lolesports"
	"jellyfin-lolesports-streams/internal/store"
	"jellyfin-lolesports-streams/internal/twitch"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//go:embed assets
var assets embed.FS

func main() {
	config := loadConfiguration()

	serverErrors := make(chan error, 1)
	httpClient := http.DefaultClient

	riotClient := lolesports.NewRiotClient(httpClient)
	twClient := twitch.NewClient(httpClient)
	twClient.SupportedCodecs = []string{"h264", "h265", "av1"}

	SQLStore, err := store.NewSqliteStore()
	if err != nil {
		panic(err)
	}

	handler := &api.Handler{
		RiotClient:   riotClient,
		TwitchClient: twClient,
		Store:        SQLStore,
		Config:       config,
	}

	srv := api.NewRouter(handler, assets)

	go func() {
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	shutdownHttpServer(srv)
}

func shutdownHttpServer(srv *http.Server) {
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Println("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}

func loadConfiguration() api.RouterConfig {
	var routerConfig api.RouterConfig

	defaults := api.RouterConfig{
		ApplicationUrl:   "http://localhost:80",
		TwitchClientId:   "kimne78kx3ncx6brgo4mv6wki5h1ko",
		DefaultUserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:157.0) Gecko/20100101 Firefox/157.0",
		CacheTTL:         15 * time.Minute,
	}

	routerConfig.ApplicationUrl = getEnv("ApplicationUrl", defaults.ApplicationUrl)
	routerConfig.TwitchClientId = getEnv("TwitchClientId", defaults.TwitchClientId)
	routerConfig.DefaultUserAgent = getEnv("DefaultUserAgent", defaults.DefaultUserAgent)
	routerConfig.CacheTTL = defaults.CacheTTL

	return routerConfig
}

func getEnv(envName string, defaultValue string) string {
	envValue := os.Getenv(envName)
	if envValue == "" {
		message := fmt.Sprintf("%s Environment Variable is empty. Using Default: %s", envName, defaultValue)
		fmt.Println(message)
		return defaultValue
	}

	return envValue
}
