package streamProvider

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

type TwitchClient struct {
	http   *http.Client
	config TwitchConfig
}

type TwitchConfig struct {
	TwitchClientId string
	UserAgent      string
	Codecs         []string
}

type playbackToken struct {
	Signature string
	Value     string
}

func NewTwitchClient(client *http.Client, config TwitchConfig) *TwitchClient {
	if client == nil {
		client = http.DefaultClient
	}
	return &TwitchClient{http: client, config: config}
}

func (c TwitchClient) postGraphQL(ctx context.Context, payload map[string]any, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://gql.twitch.tv/gql", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Client-ID", c.config.TwitchClientId)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.config.UserAgent)

	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}
	}(res.Body)

	resData, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		return fmt.Errorf("GraphQL Request Failed %s:%s", res.Status, string(resData))
	}

	return json.Unmarshal(resData, out)
}

func (c TwitchClient) GetTwitchStreamPlaylist(ctx context.Context, channel string) (string, error) {
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

func (c TwitchClient) usherUrl(channel string, extra map[string]string) string {
	var codecs string

	c.config.Codecs = append(c.config.Codecs, "h264")
	codecs = strings.Join(c.config.Codecs, ",")

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

func (c TwitchClient) accessToken(ctx context.Context, channel string) (playbackToken, error) {
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
