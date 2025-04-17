package alldebrid

import (
	"net/url"
)

type UnlockLinkDataStream struct {
	Id       string `json:"id"`
	Ext      string `json:"ext"`
	Quality  string `json:"quality,string"`
	Filesize int    `json:"filesize"`
	Proto    string `json:"proto"`
	Name     string `json:"name"`
}

type UnlockLinkData struct {
	Delayed    int                    `json:"delayed"`
	Filename   string                 `json:"filename"`
	Filesize   int                    `json:"filesize"`
	Host       string                 `json:"host"`
	HostDomain string                 `json:"hostDomain"`
	Id         string                 `json:"id"`
	Link       string                 `json:"link"`
	Paws       bool                   `json:"paws"`
	Streams    []UnlockLinkDataStream `json:"streams"`
	Path       []ResponseMagnetFile   `json:"path"`
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
