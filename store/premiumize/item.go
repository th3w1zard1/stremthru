package premiumize

import (
	"net/url"
	"time"
)

type ListItemsDataFile struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
	Size      int64  `json:"size"`
	MimeType  string `json:"mime_type"`
	VirusScan VirusScan
	Path      string
}

func (c ListItemsDataFile) GetCreatedAt() time.Time {
	return time.Unix(c.CreatedAt, 0).UTC()
}

type ListItemsData struct {
	Files []ListItemsDataFile
}

type listItemsData struct {
	ResponseContainer
	ListItemsData
}

type ListItemsParams struct {
	Ctx
}

func (c APIClient) ListItems(params *ListItemsParams) (APIResponse[ListItemsData], error) {
	response := &listItemsData{}
	res, err := c.Request("GET", "/item/listall", params, response)
	return newAPIResponse(res, response.ListItemsData), err
}

type GetItemData struct {
	Id                string    `json:"id"`
	UserId            int       `json:"user_id"`
	CustomerId        int       `json:"customer_id"`
	Name              string    `json:"name"`
	Size              int64     `json:"size"`
	CreatedAt         int64     `json:"created_at"`
	TranscodeStatus   string    `json:"transcode_status"` // pending / good_as_is
	FolderId          string    `json:"folder_id"`
	ServerName        string    `json:"server_name"`
	ACodec            string    `json:"acodec"`
	VCodec            string    `json:"vcodec"`
	MimeType          string    `json:"mime_type"`
	OpensubtitlesHash string    `json:"opensubtitles_hash"`
	ResX              string    `json:"resx"`
	ResY              string    `json:"resy"`
	Duration          string    `json:"duration"`
	VirusScan         VirusScan `json:"virus_scan"`
	AudioTrackNames   []string  `json:"audio_track_names"`
	CRC32             string    `json:"crc32"`
	Type              string    `json:"type"`
	Link              string    `json:"link"`
	DirectLink        string    `json:"directlink"`
	StreamLink        string    `json:"streamlink,omitempty"`
	Unpackable        bool      `json:"unpackable"`
}

func (d GetItemData) GetCreatedAt() time.Time {
	return time.Unix(d.CreatedAt, 0).UTC()
}

type getItemData struct {
	ResponseContainer
	GetItemData
}

type GetItemParams struct {
	Ctx
	Id string
}

func (c APIClient) GetItem(params *GetItemParams) (APIResponse[GetItemData], error) {
	params.Query = &url.Values{
		"id": []string{params.Id},
	}
	response := &getItemData{}
	res, err := c.Request("GET", "/item/details", params, response)
	return newAPIResponse(res, response.GetItemData), err
}
