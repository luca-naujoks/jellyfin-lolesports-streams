package twitch

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

var (
	nameRE   = regexp.MustCompile(`IVS-NAME="([^"]+)"`)
	sourceRE = regexp.MustCompile(`IVS-VARIANT-SOURCE="(source|transcode)"`)
)

type streamInfo struct {
	name   string
	source string
}

func parseTwitchChannelUrl(rawUrl string) (target, error) {
	url, err := url.Parse(rawUrl)
	if err != nil {
		return target{}, err
	}

	var t target
	urlSegments := extractPathSegments(url.Path)
	if len(urlSegments) > 1 || urlSegments[0] == "" {
		return target{}, errors.New("twitch url is raw no channel path parameter")
	}

	t.channel = urlSegments[0]

	return t, nil
}

func extractPathSegments(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}

	return strings.Split(path, "/")
}
