package api

import (
	"bytes"
	"embed"
	"io/fs"
	"jellyfin-lolesports-streams/internal/lolesports"
	"jellyfin-lolesports-streams/internal/store"
	"jellyfin-lolesports-streams/internal/twitch"
	"net/http"
	"path"
	"strings"
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
	Assets       embed.FS
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
		Assets: handlerConfig.Assets,
	}
}

func NewRouter(handlerConfig *Handler, webInterface embed.FS) *http.Server {
	r := gin.Default()
	gin.SetMode(gin.ReleaseMode)

	handler := NewHandler(handlerConfig)

	assetsFS, err := fs.Sub(handler.Assets, "assets")
	if err != nil {
		panic(err)
	}

	webFS, err := fs.Sub(webInterface, "web-app/dist/client")
	if err != nil {
		panic(err)
	}

	api := r.Group("/api")
	{
		api.GET("/leagues", handler.getLeagueshandler)

		api.GET("/channels", handler.GetChannelsHandler)
		api.GET("/channels/:id", handler.GetChannelHandler)
		api.POST("/channels", handler.CreateChannelHandler)
		api.PUT("/channels/:id", handler.UpdateChannelHandler)
		api.DELETE("/channels/:id", handler.DeleteChannelHandler)
		api.GET("/image/match/:id", handler.generateImage)
	}

	tv := r.Group("/tv")
	{
		tv.GET("/channels.m3u", handler.m3uHandler)
		tv.GET("/guide.xml", handler.guideXMLHandler)
		tv.GET("/stream/:channel", handler.getStreamHandler)

	}

	r.StaticFS("/static", http.FS(assetsFS))

	registerEmbeddedRoutes(r, webFS)

	srv := &http.Server{
		Addr:    ":80",
		Handler: r.Handler(),
	}

	return srv
}

func registerEmbeddedRoutes(r *gin.Engine, content fs.FS) {
	// Generic web application dist paths
	r.GET("/assets/*filepath", func(c *gin.Context) {
		serveEmbeddedFile(c.Writer, c.Request, content, "assets", c.Param("filepath"))
	})
	r.GET("/favicon.ico", func(c *gin.Context) {
		serveEmbeddedFile(c.Writer, c.Request, content, "", "favicon.svg")
	})
	r.GET("/", func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.String(), "/assets/") {
			http.NotFound(c.Writer, c.Request)
			return
		}
		serveEmbeddedFile(c.Writer, c.Request, content, "", "index.html")
	})
}

func serveEmbeddedFile(w http.ResponseWriter, r *http.Request, content fs.FS, prefix string, filepath string) {
	if filepath == "" || strings.HasSuffix(filepath, "/") {
		filepath = path.Join(filepath, "index.html")
	}

	fullPath := path.Join(prefix, filepath)

	data, err := fs.ReadFile(content, fullPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	http.ServeContent(
		w,
		r,
		path.Base(fullPath),
		time.Time{},
		bytes.NewReader(data),
	)
}
