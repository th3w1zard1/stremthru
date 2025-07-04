package alldebrid

import (
	"encoding/json"
	"net/url"
	"strconv"
)

type UnlockLinkDataStream struct {
	ABR      int     `json:"abr"`
	Ext      string  `json:"ext"`
	Filesize int     `json:"filesize"`
	Id       string  `json:"id"`
	Name     string  `json:"name"`
	Proto    string  `json:"proto"`
	Quality  string  `json:"quality,string"`
	TB       float32 `json:"tb"`
}

type UnlockLinkDataPath []ResponseMagnetFile

func (p *UnlockLinkDataPath) UnmarshalJSON(data []byte) error {
	if string(data) == `""` {
		return nil
	}
	value := []ResponseMagnetFile{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*p = value
	return nil
}

type UnlockLinkData struct {
	Delayed    int                    `json:"delayed"`
	Filename   string                 `json:"filename"`
	Filesize   int                    `json:"filesize"`
	Host       string                 `json:"host"` // magnet
	HostDomain string                 `json:"hostDomain,omitempty"`
	Id         string                 `json:"id"`
	Link       string                 `json:"link"`
	MaxChunks  int                    `json:"max_chunks,omitempty"`
	Path       UnlockLinkDataPath     `json:"path"`
	Paws       bool                   `json:"paws"`
	Streaming  []any                  `json:"streaming,omitempty"`
	Streams    []UnlockLinkDataStream `json:"streams,omitempty"`
}

func (ld UnlockLinkData) GetPath() string {
	for _, file := range getFlatFiles(ld.Path, nil, nil, 0) {
		if file.Name == ld.Filename {
			return file.Path
		}
	}
	return ""
}

type UnlockLinkParams struct {
	Ctx
	Link     string
	Password string
	UserIP   string
}

func (c APIClient) UnlockLink(params *UnlockLinkParams) (APIResponse[UnlockLinkData], error) {
	form := &url.Values{}
	form.Add("link", params.Link)
	if len(params.Password) > 0 {
		form.Add("password", params.Password)
	}
	if len(params.UserIP) > 0 {
		form.Add("userip", params.UserIP)
	}
	params.Form = form

	response := &Response[UnlockLinkData]{}
	res, err := c.Request("GET", "/v4/link/unlock", params, response)
	return newAPIResponse(res, response.Data), err
}

type GetStreamingLinkData struct {
	Link     string `json:"link"` // Optional. Download link, ONLY if available. This attribute WONT BE RETURNED if download link is a delayed link.
	Filename string `json:"filename"`
	Filesize int64  `json:"filesize"`
	Delayed  int    `json:"delayed"` // Optional. Delayed ID to get download link with delayed link flow (see next section)
}

type GetStreamingLinkParams struct {
	Ctx
	Id     string
	Stream string
}

func (c APIClient) GetStreamingLink(params *GetStreamingLinkParams) (APIResponse[GetStreamingLinkData], error) {
	params.Form = &url.Values{
		"id":     []string{params.Id},
		"stream": []string{params.Stream},
	}

	response := &Response[GetStreamingLinkData]{}
	res, err := c.Request("POST", "/v4/link/streaming", params, response)
	return newAPIResponse(res, response.Data), err
}

type DelayedLinkStatusCode int

const (
	DelayedLinkStatusCodeProcessing DelayedLinkStatusCode = 1
	DelayedLinkStatusCodeAvailable  DelayedLinkStatusCode = 2
	DelayedLinkStatusCodeError      DelayedLinkStatusCode = 3
)

type GetDelayedLinkData struct {
	Status   DelayedLinkStatusCode `json:"status"`
	TimeLeft int                   `json:"time_left"`
	Link     string                `json:"link"` // Download link, available when it is ready.
}

type GetDelayedLinkParams struct {
	Ctx
	Id int
}

func (c APIClient) GetDelayedLink(params *GetDelayedLinkParams) (APIResponse[GetDelayedLinkData], error) {
	params.Form = &url.Values{
		"id": []string{strconv.Itoa(params.Id)},
	}

	response := &Response[GetDelayedLinkData]{}
	res, err := c.Request("GET", "/v4/link/delayed", params, response)
	return newAPIResponse(res, response.Data), err
}
