package api

import (
	"embed"
	"io/fs"
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
	Store        *store.Store
	GuideCache   GuideCache
	Config       RouterConfig
}

func NewHandler(handlerConfig *Handler) *Handler {
	return &Handler{
		RiotClient:   handlerConfig.RiotClient,
		TwitchClient: handlerConfig.TwitchClient,
		Store:        handlerConfig.Store,
		GuideCache: GuideCache{
			guide: nil,
			hash:  [32]byte{},
			// Init the Cache with a definitely outdated cache time
			createdAt: time.Now().AddDate(0, 0, -1),
		},
		Config: handlerConfig.Config,
	}
}

func NewRouter(handlerConfig *Handler, assets embed.FS) *http.Server {
	assetsFS, err := fs.Sub(assets, "assets")
	if err != nil {
		panic(err)
	}

	r := gin.Default()

	//gin.SetMode(gin.ReleaseMode)

	handler := NewHandler(handlerConfig)

	r.GET("/tv/channels.m3u", handler.m3uHandler)
	r.GET("/tv/guide.xml", handler.guideXMLHandler)
	r.GET("/tv/stream/:channel", handler.getStreamHandler)

	r.GET("/api/channels", handler.GetChannelsHandler)
	r.GET("/api/channels/:id", handler.GetChannelHandler)
	r.POST("/api/channels", handler.CreateChannelHandler)
	r.PUT("/api/channels/:id", handler.UpdateChannelHandler)
	r.DELETE("/api/channels/:id", handler.DeleteChannelHandler)

	r.GET("/api/leagues", handler.getLeagueshandler)
	r.GET("/image/match/:id", handler.generateImage)

	r.StaticFS("/assets", http.FS(assetsFS))

	srv := &http.Server{
		Addr:    ":80",
		Handler: r.Handler(),
	}

	return srv
}
