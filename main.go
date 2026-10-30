package main

import (
	"context"
	"errors"
	"jellyfin-lolesports-streams/internal/api"
	"jellyfin-lolesports-streams/internal/lolesports"
	"jellyfin-lolesports-streams/internal/store"
	"jellyfin-lolesports-streams/internal/twitch"
	"jellyfin-lolesports-streams/internal/types"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var routerConfig = api.RouterConfig{
	ApplicationUrl:   "http://host.containers.internal:8080/tv/stream/",
	TwitchClientId:   "kimne78kx3ncx6brgo4mv6wki5h1ko",
	DefaultUserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:157.0) Gecko/20100101 Firefox/157.0",
	CacheTTL:         15 * time.Minute,
}

func main() {
	serverErrors := make(chan error, 1)
	httpClient := http.DefaultClient

	riotClient := lolesports.NewRiotClient(httpClient)
	twClient := twitch.NewClient(httpClient)
	twClient.SupportedCodecs = []string{"h264", "h265", "av1"}

	channelStore := store.NewStore()

	err := createDummyChannels(channelStore)

	if err != nil {
		return
	}

	handlerConfig := &api.Handler{
		RiotClient:   riotClient,
		TwitchClient: twClient,
		ChannelStore: channelStore,
		Config:       routerConfig,
	}

	srv := api.NewRouter(handlerConfig)

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

func createDummyChannels(channelStore *store.Store) error {
	err := channelStore.Add(types.Channel{
		ID:       "tolkin",
		Name:     "Tolkin",
		Provider: "twitch",
		IconUrl:  "https://static-cdn.jtvnw.net/jtv_user_pictures/54b3baeb-902e-4c18-9488-6530ada7f8ef-profile_image-70x70.png",
		Leagues: []types.League{
			{
				Id:       "98767975604431411",
				Slug:     "worlds",
				Name:     "Worlds",
				Region:   "INTERNATIONAL",
				ImageUrl: "https://static.lolesports.com/leagues/1592594612171_WorldsDarkBG.png",
				Priority: 1,
			},
			{
				Id:       "98767991310872058",
				Slug:     "lck",
				Name:     "LCK",
				Region:   "KOREA",
				ImageUrl: "https://static.lolesports.com/leagues/lck-color-on-black.png",
				Priority: 1,
			},
		},
		Language: "de",
	})
	if err != nil {
		return err
	}

	err = channelStore.Add(types.Channel{
		ID:       "riotgames",
		Name:     "Riot Games",
		Provider: "twitch",
		IconUrl:  "https://static-cdn.jtvnw.net/jtv_user_pictures/35b02a12-d516-499e-90f8-7899f136fa18-profile_image-70x70.png",
		Leagues: []types.League{
			{
				Id:       "98767975604431411",
				Slug:     "worlds",
				Name:     "Worlds",
				Region:   "INTERNATIONAL",
				ImageUrl: "https://static.lolesports.com/leagues/1592594612171_WorldsDarkBG.png",
				Priority: 1,
			},
		},
		Language: "en",
	})
	if err != nil {
		return err
	}

	return nil
}
