package xml

import (
	"encoding/xml"
	"fmt"
	"jellyfin-lolesports-streams/internal/types"
	"strings"
	"time"
)

const xmlTVTimeLayout = "20060102150405 -0700"

type xmlProgramm struct {
	XMLName xml.Name `xml:"programme"`
	Channel string   `xml:"channel,attr"`
	Start   string   `xml:"start,attr"`
	Stop    string   `xml:"stop,attr,omitempty"`

	Title       string     `xml:"title"`
	SubTitle    string     `xml:"sub-title"`
	Description string     `xml:"desc"`
	Category    string     `xml:"category"`
	Icon        *xmlIcon   `xml:"icon"`
	Images      []xmlImage `xml:"image,omitempty"`
}
type xmlChannel struct {
	XMLName     xml.Name `xml:"channel"`
	ChannelId   string   `xml:"id,attr,omitempty"`
	DisplayName string   `xml:"display-name"`
	Icon        *xmlIcon `xml:"icon"`
}

type xmlIcon struct {
	XMLName xml.Name `xml:"icon"`
	Src     string   `xml:"src,attr"`
}
type xmlImage struct {
	Type   string `xml:"type,attr"`
	Size   int    `xml:"size,attr,omitempty"`
	Orient string `xml:"orient,attr,omitempty"`
	URL    string `xml:",chardata"`
}
type xmlTV struct {
	XMLName   xml.Name      `xml:"tv"`
	Channel   []xmlChannel  `xml:"channel"`
	Programme []xmlProgramm `xml:"programme"`
}

func GenerateTVGuide(availableChannels []types.Channel, availablePrograms []types.Event, applicationUrl string) ([]byte, error) {
	var programme []xmlProgramm
	var channels []xmlChannel

	leagueChannels := make(map[string][]string)

	for _, ch := range availableChannels {
		channel := xmlChannel{
			ChannelId:   ch.ID,
			DisplayName: ch.Name,
			Icon: &xmlIcon{
				Src: ch.IconUrl,
			},
		}
		channels = append(channels, channel)

		for _, league := range ch.Leagues {
			leagueChannels[league.Slug] = append(leagueChannels[league.Slug], ch.ID)
		}
	}

	for _, p := range availablePrograms {
		programChannels := leagueChannels[p.League]

		for _, ch := range programChannels {
			program := xmlProgramm{
				Channel:     ch,
				Start:       p.StartTime.Format(xmlTVTimeLayout),
				Stop:        p.StartTime.Add(time.Duration(p.BestOf) * time.Hour).Format(xmlTVTimeLayout),
				Title:       fmt.Sprintf("%s vs %s", p.Teams[0].Code, p.Teams[1].Code),
				SubTitle:    fmt.Sprintf("%s - %s - Best of %v", p.League, p.Stage, p.BestOf),
				Description: fmt.Sprintf("%s vs %s", p.Teams[0].Name, p.Teams[1].Name),
				Category:    "League of Legends",
				Icon: &xmlIcon{
					Src: fmt.Sprintf("%s/image/match/%s", applicationUrl, p.ID),
				},
				Images: []xmlImage{
					{
						Type:   "backdrop",
						Size:   3,
						Orient: "L",
						URL:    fmt.Sprintf("%s/assets/594917.jpg", applicationUrl),
					}, {
						Type:   "still",
						Size:   3,
						Orient: "L",
						URL:    fmt.Sprintf("%s/image/match/%s", applicationUrl, p.ID),
					},
				},
			}
			programme = append(programme, program)
		}
	}

	tv := xmlTV{
		Programme: programme,
		Channel:   channels,
	}

	data, err := xml.MarshalIndent(tv, "", "   ")
	if err != nil {
		return nil, err
	}

	return data, err
}

func GenerateM3u(channels []types.Channel, baseUrl string) []byte {
	builder := strings.Builder{}

	builder.WriteString("#EXTM3U\n")

	for _, channel := range channels {
		extInf := fmt.Sprintf("#EXTINF:-1 tvg-id=\"%s\" tvg-logo=\"%s\" group-title=\"Esports\",%s\n", channel.ID, channel.IconUrl, channel.Name)
		builder.WriteString(extInf)
		streamUrl := fmt.Sprintf("%s%s\n", baseUrl, channel.ID)
		builder.WriteString(streamUrl)
	}

	return []byte(builder.String())
}
