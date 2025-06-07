package torznab

import (
	"encoding/xml"
	"strconv"
	"strings"
	"time"

	"github.com/MunifTanjim/stremthru/core"
)

const rfc822 = "Mon, 02 Jan 2006 15:04:05 -0700"

type ChannelItemEnclosure struct {
	XMLName xml.Name `xml:"enclosure"`
	URL     string   `xml:"url,attr,omitempty"`
	Length  int64    `xml:"length,attr,omitempty"`
	Type    string   `xml:"type,attr,omitempty"`
}

type ChannelItemAttribute struct {
	XMLName xml.Name `xml:"torznab:attr"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:"value,attr"`
}

type ChannelItem struct {
	XMLName xml.Name `xml:"item"`

	// standard rss elements
	Category    string               `xml:"category,omitempty"`
	Description string               `xml:"description,omitempty"`
	Enclosure   ChannelItemEnclosure `xml:"enclosure,omitempty"`
	Files       int                  `xml:"files,omitempty"`
	GUID        string               `xml:"guid,omitempty"`
	Link        string               `xml:"link,omitempty"`
	PublishDate string               `xml:"pubDate,omitempty"`
	Title       string               `xml:"title,omitempty"`

	Attributes []ChannelItemAttribute
}

type Channel struct {
	XMLName     xml.Name `xml:"channel"`
	Title       string   `xml:"title,omitempty"`
	Description string   `xml:"description,omitempty"`
	Link        string   `xml:"link,omitempty"`
	Language    string   `xml:"language,omitempty"`
	Category    string   `xml:"category,omitempty"`
	Items       []ResultItem
}

type RSS struct {
	XMLName          xml.Name `xml:"rss"`
	AtomNamespace    string   `xml:"xmlns:atom,attr"`
	TorznabNamespace string   `xml:"xmlns:torznab,attr"`
	Version          string   `xml:"version,attr,omitempty"`
	Channel          Channel  `xml:"channel"`
}

type ResultItem struct {
	Category    Category
	Description string
	Files       int
	GUID        string
	Link        string
	PublishDate time.Time
	Title       string

	Audio      string
	Codec      string
	IMDB       string
	InfoHash   string
	Language   string
	Resolution string
	Site       string
	Size       int64
	Year       int
}

func (ri ResultItem) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	attrs := []ChannelItemAttribute{}
	if ri.Audio != "" {
		attrs = append(attrs, ChannelItemAttribute{Name: "audio", Value: ri.Audio})
	}
	attrs = append(attrs, ChannelItemAttribute{Name: "category", Value: strconv.Itoa(ri.Category.ID)})
	if ri.IMDB != "" {
		attrs = append(attrs, ChannelItemAttribute{Name: "imdb", Value: strings.TrimPrefix(ri.IMDB, "tt")})
	}
	if ri.InfoHash != "" {
		if magnet, err := core.ParseMagnetLink(ri.InfoHash); err == nil {
			attrs = append(
				attrs,
				ChannelItemAttribute{Name: "infohash", Value: magnet.Hash},
				ChannelItemAttribute{Name: "magneturl", Value: magnet.Link},
			)
			if ri.GUID == "" {
				ri.GUID = magnet.Hash
			}
			if ri.Link == "" {
				ri.Link = magnet.Link
			}
		}
	}
	if ri.Language != "" {
		attrs = append(attrs, ChannelItemAttribute{Name: "language", Value: ri.Language})
	}
	if ri.Resolution != "" {
		attrs = append(attrs, ChannelItemAttribute{Name: "resolution", Value: ri.Resolution})
	}
	if ri.Site != "" {
		attrs = append(attrs, ChannelItemAttribute{Name: "site", Value: ri.Site})
	}
	if ri.Size > 0 {
		attrs = append(attrs, ChannelItemAttribute{Name: "size", Value: strconv.FormatInt(ri.Size, 10)})
	}
	if ri.Codec != "" {
		attrs = append(attrs, ChannelItemAttribute{Name: "video", Value: ri.Codec})
	}
	if ri.Year != 0 {
		attrs = append(attrs, ChannelItemAttribute{Name: "year", Value: strconv.Itoa(ri.Year)})
	}
	return e.Encode(ChannelItem{
		Attributes:  attrs,
		Category:    ri.Category.Name,
		Description: ri.Description,
		Files:       ri.Files,
		GUID:        ri.GUID,
		Link:        ri.Link,
		PublishDate: ri.PublishDate.Format(rfc822),
		Title:       ri.Title,
		Enclosure: ChannelItemEnclosure{
			URL:    ri.Link,
			Length: ri.Size,
			Type:   "application/x-bittorrent;x-scheme-handler/magnet",
		},
	})
}

type ResultFeed struct {
	Info  Info
	Items []ResultItem
}

func (rf ResultFeed) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.Encode(RSS{
		Version: "2.0",
		Channel: Channel{
			Category:    rf.Info.Category,
			Description: rf.Info.Description,
			Items:       rf.Items,
			Language:    rf.Info.Language,
			Link:        rf.Info.Link,
			Title:       rf.Info.Title,
		},
		AtomNamespace:    "http://www.w3.org/2005/Atom",
		TorznabNamespace: "http://torznab.com/schemas/2015/feed",
	})
}
