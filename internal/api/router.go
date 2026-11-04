package api

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"io/fs"
	"jellyfin-lolesports-streams/internal/lolesports"
	"jellyfin-lolesports-streams/internal/store"
	"jellyfin-lolesports-streams/internal/streamProvider"
	"jellyfin-lolesports-streams/internal/types"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	*http.Server
}

type RouterConfig struct {
	ApplicationUrl string
	CacheTTL       time.Duration
	DevMode        bool
}

type Handler struct {
	RiotClient   *lolesports.Client
	TwitchClient *streamProvider.TwitchClient
	Store        *store.Store
	GuideCache   GuideCache
	Config       RouterConfig
	Assets       embed.FS
}

func (s *Server) Run(c chan<- error) {
	err := s.Server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		c <- err
	}
}

func (s *Server) ShutdownHttpServer() {
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.Server.Shutdown(ctx); err != nil {
		log.Println("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}

func NewHandler(httpClient *http.Client, config types.MainConfig, sqlStore *store.Store, assets embed.FS) *Handler {
	return &Handler{
		RiotClient: lolesports.NewRiotClient(httpClient, lolesports.Config{
			BaseUrl: config.LolEsportsApiUrl,
			Secret:  config.LolEsportsSecret,
		}),
		TwitchClient: streamProvider.NewTwitchClient(httpClient, streamProvider.TwitchConfig{
			TwitchClientId: config.TwitchClientId,
			UserAgent:      config.TwitchUserAgent,
			Codecs:         []string{"h264", "h265", "av1"},
		}),
		Store: sqlStore,
		Config: RouterConfig{
			ApplicationUrl: config.ApplicationUrl,
			CacheTTL:       config.CacheTTL,
		},
		Assets: assets,
	}
}

func NewRouter(handler *Handler, webInterface embed.FS) *Server {
	if !handler.Config.DevMode {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

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
		api.GET("/leagues", handler.getLeaguesHandler)

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

	srv := &Server{
		Server: &http.Server{
			Addr:    ":80",
			Handler: r.Handler(),
		},
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
