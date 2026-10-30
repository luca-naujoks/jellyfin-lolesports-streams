package twitch

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
)

const (
	twitchClientID   = "kimne78kx3ncx6brgo4mv6wki5h1ko"
	defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:157.0) Gecko/20100101 Firefox/157.0"
	defaultCodec     = "h264"
)

type Client struct {
	http            *http.Client
	SupportedCodecs []string
}

type Stream struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Url  string `json:"url"`
}

type target struct {
	channel string
}
type playbackToken struct {
	Signature string
	Value     string
}

func NewClient(client *http.Client) *Client {
	if client == nil {
		client = http.DefaultClient
	}
	return &Client{http: client}
}

func (c Client) Streams(ctx context.Context, rawURL string) (string, error) {
	target, err := parseTwitchChannelUrl(rawURL)
	if err != nil {
		return "", err
	}
	switch true {
	case target.channel != "":
		return c.liveStreams(ctx, target.channel)
	default:
		return "", errors.New(fmt.Sprintf("target.channel: %s, this shouldn't be empty", target.channel))
	}
}

func (c Client) postGraphQL(ctx context.Context, payload map[string]any, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://gql.twitch.tv/gql", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Client-ID", twitchClientID)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", defaultUserAgent)

	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	resData, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		return fmt.Errorf("GraphQL Request Failed %s:%s", res.Status, string(resData))
	}

	return json.Unmarshal(resData, out)
}

func (c Client) liveStreams(ctx context.Context, channel string) (string, error) {
	token, err := c.accessToken(ctx, channel)
	if err != nil {
		return "", err
	}

	playlistURL := c.usherUrl(channel,
		map[string]string{
			"sig":        token.Signature,
			"token":      token.Value,
			"fast_bread": "true",
		})

	return playlistURL, err
}

func (c Client) usherUrl(channel string, extra map[string]string) string {
	var codecs string
	if len(c.SupportedCodecs) > 0 {
		c.SupportedCodecs = append(c.SupportedCodecs, "h264")
		codecs = strings.Join(c.SupportedCodecs, ",")
	} else {
		codecs = defaultCodec
	}

	values := url.Values{
		"platform":                   {"web"},
		"p":                          {fmt.Sprint(rand.Intn(999999))},
		"allow_source":               {"true"},
		"allow_audio_only":           {"true"},
		"playlist_include_framerate": {"true"},
		"multigroup_video":           {"true"},
		"supported_codecs":           {codecs},
	}

	for key, value := range extra {
		values.Set(key, value)
	}

	return (&url.URL{
		Scheme:   "https",
		Host:     "usher.ttvnw.net",
		Path:     fmt.Sprintf("/api/v2/channel/hls/%s.m3u8", channel),
		RawQuery: values.Encode(),
	}).String()
}

func (c Client) accessToken(ctx context.Context, channel string) (playbackToken, error) {
	variables := map[string]any{
		"isLive":     true,
		"login":      channel,
		"isVod":      false,
		"vodID":      "",
		"playerType": "embed",
		"platform":   "site",
	}

	payload := map[string]any{
		"operationName": "PlaybackAccessToken",
		"extensions": map[string]any{
			"persistedQuery": map[string]any{
				"version":    1,
				"sha256Hash": "ed230aa1e33e07eebb8928504583da78a5173989fadfb1ac94be06a04f3cdbe9",
			},
		},
		"variables": variables,
	}

	var response struct {
		Data struct {
			Stream *struct {
				Value     string `json:"value"`
				Signature string `json:"signature"`
			} `json:"streamPlaybackAccessToken"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
		Error   string `json:"error"`
		Message string `json:"message"`
	}

	err := c.postGraphQL(ctx, payload, &response)
	if err != nil {
		return playbackToken{}, err
	}
	if len(response.Errors) > 0 {
		return playbackToken{}, errors.New(response.Errors[0].Message)
	}
	if response.Error != "" || response.Message != "" {
		if response.Error != "" && response.Message != "" {
			return playbackToken{}, fmt.Errorf("%s: %s", response.Error, response.Message)
		}
		return playbackToken{}, errors.New(response.Error + response.Message)
	}

	var signature, value string
	if response.Data.Stream != nil {
		signature, value = response.Data.Stream.Signature, response.Data.Stream.Value
	}

	if signature == "" || value == "" {
		return playbackToken{}, errors.New("playback token was not returned")
	}

	return playbackToken{
		Signature: signature,
		Value:     value,
	}, nil

}
