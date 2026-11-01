package api

import (
	"context"
	"fmt"
	"jellyfin-lolesports-streams/internal/lolesports"
	"jellyfin-lolesports-streams/internal/types"
	"jellyfin-lolesports-streams/internal/xml"
	"maps"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

func (h *Handler) guideXMLHandler(c *gin.Context) {
	leagueChannels := h.Store.GetAll()

	if len(leagueChannels) <= 0 {
		// INFO no channels therefore no League Channels nor Program to generate
		c.Data(http.StatusOK, "text/xml", []byte(""))
		return
	}

	uniqueLeagueIds := make(map[string]struct{})

	for i := range leagueChannels {
		for _, league := range leagueChannels[i].Leagues {
			uniqueLeagueIds[league.Id] = struct{}{}
		}
	}

	leagueIds := slices.Sorted(maps.Keys(uniqueLeagueIds))

	if guide, ok := h.GuideCache.Get(leagueChannels, h.Config.CacheTTL); ok {
		c.Data(http.StatusOK, "text/xml", guide)
		return
	}

	events, err := h.RiotClient.RequestLeagueSchedule(leagueIds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	h.Store.Events = events

	guide, err := xml.GenerateTVGuide(leagueChannels, events, h.Config.ApplicationUrl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	h.GuideCache.Set(guide, leagueChannels)
	c.Data(http.StatusOK, "text/xml", guide)
}

func (h *Handler) m3uHandler(c *gin.Context) {
	channels := h.Store.GetAll()

	m3u := xml.GenerateM3u(channels, fmt.Sprintf("%s/tv/stream/", h.Config.ApplicationUrl))

	c.Data(
		http.StatusOK,
		"audio/x-mpegurl",
		m3u,
	)
}

func (h *Handler) getStreamHandler(c *gin.Context) {
	channel := c.Param("channel")

	channels := h.Store.GetAll()

	idx := slices.IndexFunc(channels, func(ch types.Channel) bool {
		return ch.ID == channel
	})

	if idx < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("No Channel Found with id %s", channel)})
		return
	}

	twitchUrl := fmt.Sprintf("https://twitch.tv/%s", channel)

	playlistUrl, err := h.TwitchClient.Streams(context.Background(), twitchUrl)

	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.Redirect(http.StatusFound, playlistUrl)
}

func (h *Handler) GetChannelsHandler(c *gin.Context) {
	channels := h.Store.GetAll()

	c.JSON(http.StatusOK, channels)
}
func (h *Handler) GetChannelHandler(c *gin.Context) {
	channelId := c.Param("id")

	channel, err := h.Store.GetByID(channelId)
	if err != nil {
		c.JSON(http.StatusNotFound, err.Error())
		return
	}

	c.JSON(http.StatusOK, channel)
}

func (h *Handler) CreateChannelHandler(c *gin.Context) {
	var channel types.Channel
	err := c.ShouldBind(&channel)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	err = h.Store.Add(channel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusCreated, channel)
}
func (h *Handler) UpdateChannelHandler(c *gin.Context) {
	var channel types.Channel
	err := c.ShouldBind(&channel)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	err = h.Store.Update(channel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}
func (h *Handler) DeleteChannelHandler(c *gin.Context) {
	id := c.Param("id")

	err := h.Store.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusNoContent, gin.H{})
}

func (h *Handler) getLeagueshandler(c *gin.Context) {
	leagues, err := h.RiotClient.RequestAllLeagues()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, leagues)
}

func (h *Handler) generateImage(c *gin.Context) {
	matchId := c.Param("id")

	if h.Store.Events == nil {
		leagueChannels := h.Store.GetAll()

		if len(leagueChannels) <= 0 {
			// INFO no channels therefore no League Channels nor Program to generate
			return
		}

		uniqueLeagueIds := make(map[string]struct{})

		for i := range leagueChannels {
			for _, league := range leagueChannels[i].Leagues {
				uniqueLeagueIds[league.Id] = struct{}{}
			}
		}

		leagueIds := slices.Sorted(maps.Keys(uniqueLeagueIds))

		events, err := h.RiotClient.RequestLeagueSchedule(leagueIds)
		if err != nil {
			return
		}

		h.Store.Events = events
	}

	idx := slices.IndexFunc(h.Store.Events, func(event types.Event) bool {
		return event.ID == matchId
	})

	if idx < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no Match Found with the provided Id"})
		return
	}

	event := h.Store.Events[idx]

	image, err := lolesports.GenerateVsImage(event.Teams)
	if err != nil {
		return
	}

	c.Data(http.StatusOK, "image/png", image)
}
