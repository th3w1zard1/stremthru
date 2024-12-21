package stremio

type StreamBehaviorHintsProxyHeaders struct {
	Request  map[string]string `json:"request,omitempty"`
	Response map[string]string `json:"response,omitempty"`
}

type StreamBehaviorHints struct {
	CountryWhitelist []string                         `json:"countryWhitelist,omitempty"`
	NotWebReady      bool                             `json:"notWebReady,omitempty"`
	BingeGroup       string                           `json:"bingeGroup,omitempty"`
	ProxyHeaders     *StreamBehaviorHintsProxyHeaders `json:"proxyHeaders,omitempty"`
	VideoHash        string                           `json:"videoHash,omitempty"`
	VideoSize        int64                            `json:"videoSize,omitempty"`
	Filename         string                           `json:"filename,omitempty"`
}

type Stream struct {
	URL         string `json:"url,omitempty"`
	YoutubeID   string `json:"ytId,omitempty"`
	InfoHash    string `json:"infoHash,omitempty"`
	FileIndex   int    `json:"fileIdx,omitempty"`
	ExternalURL string `json:"externalUrl,omitempty"`

	Name          string               `json:"name,omitempty"`
	Title         string               `json:"title,omitempty"` // warning: this will soon be deprecated in favor of `description`
	Description   string               `json:"description,omitempty"`
	Subtitles     []Subtitle           `json:"subtitles,omitempty"`
	Sources       []string             `json:"sources,omitempty"`
	BehaviorHints *StreamBehaviorHints `json:"behaviorHints,omitempty"`
}
