package mdblist

import (
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/MunifTanjim/stremthru/core"
)

type Item struct {
	Id             int       `json:"id"`
	Rank           int       `json:"rank"`
	Adult          int       `json:"adult"` // 0 / 1
	Genre          []Genre   `json:"genre,omitempty"`
	Title          string    `json:"title"`
	Poster         string    `json:"poster,omitempty"`
	ImdbId         string    `json:"imdb_id"`
	TvdbId         int       `json:"tvdb_id"`
	Language       string    `json:"language"`
	Mediatype      MediaType `json:"mediatype"` // movie / show
	ReleaseYear    int       `json:"release_year"`
	SpokenLanguage string    `json:"spoken_language"`
}

type FetchListItemsData = []Item

type listResponseData[T any] struct {
	ResponseContainer
	data []T
}

func (d *listResponseData[T]) UnmarshalJSON(data []byte) error {
	var rerr ResponseContainer

	if err := json.Unmarshal(data, &rerr); err == nil {
		d.ResponseContainer = rerr
		return nil
	}

	var items []T
	if err := json.Unmarshal(data, &items); err == nil {
		d.data = items
		return nil
	}

	return core.NewAPIError("failed to parse response")
}

type PageSort = string  // rank / score / usort / score_average / released / releasedigital / imdbrating / imdbvotes / last_air_date / imdbpopular / tmdbpopular / rogerebert / rtomatoes / rtaudience / metacritic / myanimelist / letterrating / lettervotes / budget / revenue / runtime / title / added / random
type PageOrder = string // asc / desc

type FetchListItemsParams struct {
	Ctx
	ListId      int
	Limit       int
	Offset      int
	FilterGenre Genre
	Sort        PageSort
	Order       PageOrder // asc / desc
}

func (c APIClient) FetchListItems(params *FetchListItemsParams) (APIResponse[FetchListItemsData], error) {
	query := url.Values{}
	if params.Limit != 0 {
		query.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Offset != 0 {
		query.Set("offset", strconv.Itoa(params.Offset))
	}
	query.Set("append_to_response", "genre,poster")
	if params.Sort != "" {
		query.Set("sort", params.Sort)
	}
	if params.Order != "" {
		query.Set("order", params.Order)
	}
	query.Set("unified", "true")
	params.Query = &query

	response := &listResponseData[Item]{}
	res, err := c.Request("GET", "/lists/"+strconv.Itoa(params.ListId)+"/items", params, response)
	return newAPIResponse(res, response.data), err
}
