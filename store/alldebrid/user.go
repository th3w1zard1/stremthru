package alldebrid

import (
	"encoding/json"
	"time"
)

type GetUserDataUser struct {
	Username       string `json:"username"`
	Email          string `json:"email"`
	IsPremium      bool   `json:"isPremium"`
	IsSubscribed   bool   `json:"isSubscribed"`
	IsTrial        bool   `json:"isTrial"`
	PremiumUntil   int    `json:"premiumUntil"`
	Lang           string `json:"lang"`
	PreferedDomain string `json:"preferedDomain"`
	FidelityPoints int    `json:"fidelityPoints"`
}

type GetUserData struct {
	User GetUserDataUser `json:"user"`
}

type GetUserParams struct {
	Ctx
}

func (c APIClient) GetUser(params *GetUserParams) (APIResponse[GetUserDataUser], error) {
	response := &Response[GetUserData]{}
	res, err := c.Request("GET", "/v4/user", params, response)
	return newAPIResponse(res, response.Data.User), err
}

type UserLink struct {
	Link     string      `json:"link"`
	LinkDL   string      `json:"link_dl"`
	Filename string      `json:"filename"`
	Size     json.Number `json:"size"`
	Date     int64       `json:"date"`
	Host     string      `json:"host"` // `error`
}

func (l UserLink) GetSize() int64 {
	size, err := l.Size.Int64()
	if err != nil {
		size = -1
	}
	return size
}

func (l UserLink) GetDate() time.Time {
	return time.Unix(l.Date, 0).UTC()
}

type GetRecentUserLinksData struct {
	Links []UserLink `json:"links"`
}

type GetRecentUserLinksParams struct {
	Ctx
}

func (c APIClient) GetRecentUserLinks(params *GetRecentUserLinksParams) (APIResponse[[]UserLink], error) {
	response := &Response[GetRecentUserLinksData]{}
	res, err := c.Request("GET", "/v4/user/history", params, response)
	return newAPIResponse(res, response.Data.Links), err
}

type GetSavedUserLinks struct {
	Links []UserLink `json:"links"`
}

type GetSavedUserLinksParams struct {
	Ctx
}

func (c APIClient) GetSavedUserLinks(params *GetSavedUserLinksParams) (APIResponse[[]UserLink], error) {
	response := &Response[GetSavedUserLinks]{}
	res, err := c.Request("GET", "/v4/user/links", params, response)
	return newAPIResponse(res, response.Data.Links), err
}
