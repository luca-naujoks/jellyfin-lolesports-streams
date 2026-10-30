package api

import (
	"jellyfin-lolesports-streams/internal/lolesports"
	"jellyfin-lolesports-streams/internal/store"
	"jellyfin-lolesports-streams/internal/twitch"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type RouterConfig struct {
	ApplicationUrl   string
	TwitchClientId   string
	DefaultUserAgent string
	CacheTTL         time.Duration
}

type Handler struct {
	RiotClient   *lolesports.Client
	TwitchClient *twitch.Client
	ChannelStore *store.Store
	GuideCache   GuideCache
	Config       RouterConfig
}

func NewHandler(handlerConfig *Handler) *Handler {
	return &Handler{
		RiotClient:   handlerConfig.RiotClient,
		TwitchClient: handlerConfig.TwitchClient,
		ChannelStore: handlerConfig.ChannelStore,
		GuideCache: GuideCache{
			guide: nil,
			hash:  [32]byte{},
			// Init the Cache with a definitely outdated cache time
			createdAt: time.Now().AddDate(0, 0, -1),
		},
		Config: handlerConfig.Config,
	}
}

func NewRouter(handlerConfig *Handler) *http.Server {
	r := gin.Default()

	handler := NewHandler(handlerConfig)

	r.GET("/tv/channels.m3u", handler.m3uHandler)
	r.GET("/tv/guide.xml", handler.guideXMLHandler)
	r.GET("/tv/stream/:channel", handler.getStreamHandler)

	r.GET("/api/channels", handler.GetChannelsHandler)
	r.GET("/api/channels/:id", handler.GetChannelHandler)
	r.PUT("/api/channels/:id", handler.UpdateChannelHandler)
	r.DELETE("/api/channels/:id", handler.DeleteChannelHandler)

	r.GET("/api/leagues", handler.getLeagueshandler)

	srv := &http.Server{
		Addr:    ":80",
		Handler: r.Handler(),
	}

	return srv
}
