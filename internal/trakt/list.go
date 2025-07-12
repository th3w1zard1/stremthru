package trakt

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/MunifTanjim/stremthru/core"
)

type ListPrivacy = string

const (
	ListPrivacyPublic  ListPrivacy = "public"
	ListPrivacyFriends ListPrivacy = "friends"
	ListPrivacyLink    ListPrivacy = "link"
	ListPrivacyPrivate ListPrivacy = "private"
)

type List struct {
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Privacy        string    `json:"privacy"` // public / friends / link / private
	ShareLink      string    `json:"share_link"`
	Type           string    `json:"type"` // personal
	DisplayNumbers bool      `json:"display_numbers"`
	AllowComments  bool      `json:"allow_comments"`
	SortBy         string    `json:"sort_by"`  // rank
	SortHow        string    `json:"sort_how"` // asc
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	ItemCount      int       `json:"item_count"`
	CommentCount   int       `json:"comment_count"`
	Likes          int       `json:"likes"`
	Ids            struct {
		Trakt int    `json:"trakt"`
		Slug  string `json:"slug"`
	} `json:"ids"`
	User struct {
		Username string `json:"username"`
		Private  bool   `json:"private"`
		Name     string `json:"name"`
		Vip      bool   `json:"vip"`
		VipEp    bool   `json:"vip_ep"`
		Ids      struct {
			Slug string `json:"slug"`
		} `json:"ids"`
	} `json:"user"`
}

type FetchListData struct {
	ResponseError
	List
}

type FetchListParams struct {
	Ctx
	ListId int
}

func (c APIClient) FetchList(params *FetchListParams) (APIResponse[List], error) {
	response := FetchListData{}
	path := "/lists/" + strconv.Itoa(params.ListId)
	res, err := c.Request("GET", path, params, &response)
	return newAPIResponse(res, response.List), err
}

type ListItemIds struct {
	Trakt  int    `json:"trakt"`
	Slug   string `json:"slug"`
	IMDB   string `json:"imdb,omitempty"`
	TMDB   int    `json:"tmdb,omitempty"`
	TVDB   int    `json:"tvdb,omitempty"`
	TVRage any    `json:"tv_rage,omitempty"`
}

type listItemCommon struct {
	Title                 string      `json:"title"`
	Year                  int         `json:"year"`
	Ids                   ListItemIds `json:"ids"`
	Tagline               string      `json:"tagline,omitempty"`
	Overview              string      `json:"overview,omitempty"`
	Runtime               int         `json:"runtime,omitempty"` // in minutes
	Certification         string      `json:"certification,omitempty"`
	Country               string      `json:"country,omitempty"` // us
	Status                string      `json:"status,omitempty"`  // released
	Rating                float32     `json:"rating,omitempty"`  // 0.0 - 10.0
	Votes                 int         `json:"votes,omitempty"`
	CommentCount          int         `json:"comment_count,omitempty"`
	Trailer               string      `json:"trailer,omitempty"`
	Homepage              string      `json:"homepage,omitempty"`
	UpdatedAt             *time.Time  `json:"updated_at,omitempty"`
	Language              string      `json:"language,omitempty"`
	Languages             []string    `json:"languages,omitempty"`
	AvailableTranslations []string    `json:"available_translations,omitempty"`
	Genres                []string    `json:"genres,omitempty"`
	OriginalTitle         string      `json:"original_title,omitempty"`
	Images                *struct {
		Fanart   []string `json:"fanart"`
		Poster   []string `json:"poster"`
		Logo     []string `json:"logo"`
		Clearart []string `json:"clearart"`
		Banner   []string `json:"banner"`
		Thumb    []string `json:"thumb"`
	} `json:"images,omitempty"`
}

type ListItemMovie struct {
	listItemCommon
	Released      string `json:"released,omitempty"` // YYYY-MM-DD
	AfterCredits  bool   `json:"after_credits,omitempty"`
	DuringCredits bool   `json:"during_credits,omitempty"`
}

type ListItemShow struct {
	listItemCommon
	FirstAired *time.Time `json:"first_aired,omitempty"`
	Airs       *struct {
		Day      string `json:"day"`      // e.g., "Monday"
		Time     string `json:"time"`     // e.g., "20:00"
		Timezone string `json:"timezone"` // e.g., "America/New_York"
	} `json:"airs,omitempty"`
	Network       string `json:"network,omitempty"`
	AiredEpisodes int    `json:"aired_episodes,omitempty"`
}

type ItemType = string

const (
	ItemTypeMovie   ItemType = "movie"
	ItemTypeShow    ItemType = "show"
	ItemTypeSeason  ItemType = "season"
	ItemTypeEpisode ItemType = "episode"
)

type ListItem struct {
	Rank     int            `json:"rank"`
	Id       int64          `json:"id"`
	ListedAt time.Time      `json:"listed_at"`
	Note     string         `json:"note,omitempty"`
	Type     ItemType       `json:"type"`
	Movie    *ListItemMovie `json:"movie,omitempty"`
	Show     *ListItemShow  `json:"show,omitempty"`
}

type FetchListItemsData = []ListItem

type listResponseData[T any] struct {
	ResponseError
	data []T
}

func (d *listResponseData[T]) UnmarshalJSON(data []byte) error {
	var rerr ResponseError

	if err := json.Unmarshal(data, &rerr); err == nil {
		d.ResponseError = rerr
		return nil
	}

	var items []T
	err := json.Unmarshal(data, &items)
	if err == nil {
		d.data = items
		return nil
	}

	e := core.NewAPIError("failed to parse response")
	e.Cause = err
	return e
}

type FetchListItemsParams struct {
	Ctx
	ListId   int
	Type     []ItemType
	SortBy   string // rank / added / title / released / runtime / popularity / random / percentage / my_rating / watched / collected
	SortHow  string // asc / desc
	Extended string // images / full / full,images
}

func (c APIClient) FetchListItems(params *FetchListItemsParams) (APIResponse[FetchListItemsData], error) {
	params.Query = &url.Values{}
	if params.Extended != "" {
		params.Query.Set("extended", params.Extended)
	}

	response := listResponseData[ListItem]{}
	path := "/lists/" + strconv.Itoa(params.ListId) + "/items"
	if len(params.Type) > 0 {
		path += "/" + strings.Join(params.Type, ",")
	} else if params.SortBy != "" {
		path += "/movie,show"
	}
	if params.SortBy != "" {
		path += "/" + params.SortBy
		if params.SortHow != "" {
			path += "/" + params.SortHow
		}
	}
	res, err := c.Request("GET", path, params, &response)
	return newAPIResponse(res, response.data), err
}

type dynamicListMeta struct {
	Endpoint string
	Id       string
	ItemType ItemType
	Name     string
	NoLimit  bool
	NoPage   bool

	HasPeriod bool
	Period    string

	HasUserId bool
	UserId    string
}

var dynamicListMetaById = map[string]dynamicListMeta{
	"shows/trending": {
		Endpoint: "/shows/trending",
		Name:     "Trending",
		ItemType: ItemTypeShow,
	},
	"shows/anticipated": {
		Endpoint: "/shows/anticipated",
		Name:     "Anticipated",
		ItemType: ItemTypeShow,
	},
	"shows/popular": {
		Endpoint: "/shows/popular",
		Name:     "Popular",
		ItemType: ItemTypeShow,
	},
	"shows/favorited": {
		Endpoint:  "/shows/favorited/{period}",
		HasPeriod: true,
		Name:      "Most Favorited",
		ItemType:  ItemTypeShow,
	},
	"shows/watched": {
		Endpoint:  "/shows/watched/{period}",
		HasPeriod: true,
		Name:      "Most Watched",
		ItemType:  ItemTypeShow,
	},
	"shows/collected": {
		Endpoint:  "/shows/collected/{period}",
		HasPeriod: true,
		Name:      "Most Collected",
		ItemType:  ItemTypeShow,
	},
	"shows/recommendations": {
		Endpoint: "/recommendations/shows",
		NoPage:   true,
		Name:     "Recommended",
		ItemType: ItemTypeShow,
		Id:       USER_SHOWS_RECOMMENDATIONS_ID,
	},

	"movies/trending": {
		Endpoint: "/movies/trending",
		Name:     "Trending",
		ItemType: ItemTypeMovie,
	},
	"movies/anticipated": {
		Endpoint: "/movies/anticipated",
		Name:     "Anticipated",
		ItemType: ItemTypeMovie,
	},
	"movies/popular": {
		Endpoint: "/movies/popular",
		Name:     "Popular",
		ItemType: ItemTypeMovie,
	},
	"movies/favorited": {
		Endpoint:  "/movies/favorited/{period}",
		HasPeriod: true,
		Name:      "Most Favorited",
		ItemType:  ItemTypeMovie,
	},
	"movies/watched": {
		Endpoint:  "/movies/watched/{period}",
		HasPeriod: true,
		Name:      "Most Watched",
		ItemType:  ItemTypeMovie,
	},
	"movies/collected": {
		Endpoint:  "/movies/collected/{period}",
		HasPeriod: true,
		Name:      "Most Collected",
		ItemType:  ItemTypeMovie,
	},
	"movies/boxoffice": {
		Endpoint: "/movies/boxoffice",
		NoPage:   true,
		Name:     "Weekend Box Office",
		ItemType: ItemTypeMovie,
	},
	"movies/recommendations": {
		Endpoint: "/recommendations/movies",
		NoPage:   true,
		Name:     "Recommended",
		ItemType: ItemTypeMovie,
		Id:       USER_MOVIES_RECOMMENDATIONS_ID,
	},

	"favorites": {
		Endpoint:  "/users/{user_id}/favorites",
		NoPage:    true,
		NoLimit:   true,
		Name:      "Favorites",
		HasUserId: true,
	},
	"watchlist": {
		Endpoint:  "/users/{user_id}/watchlist",
		NoPage:    true,
		NoLimit:   true,
		Name:      "Watchlist",
		HasUserId: true,
	},
}

type FetchMovieRecommendationData []ListItemMovie

func GetDynamicListMeta(id string) *dynamicListMeta {
	id = strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(id, "~:"), "u:"), "/")

	if strings.Contains(id, ":") {
		parts := strings.Split(id, ":")

		meta, ok := dynamicListMetaById[parts[0]]
		if !ok {
			return nil
		}

		if meta.HasUserId {
			if len(parts) < 2 {
				return nil
			}
			meta.UserId = parts[1]
		}

		return &meta
	}

	parts := strings.Split(id, "/")
	if len(parts) < 2 || len(parts) > 3 {
		return nil
	}
	meta, ok := dynamicListMetaById[parts[0]+"/"+parts[1]]
	if !ok {
		return nil
	}
	if meta.HasPeriod {
		if len(parts) < 3 {
			parts[2] = "weekly"
		}
		meta.Period = parts[2]
		switch meta.Period {
		case "daily", "weekly", "monthly":
			meta.Name += " (" + strings.ToUpper(meta.Period[:1]) + meta.Period[1:] + ")"
		case "all":
			meta.Name += " (All Time)"
		default:
			return nil
		}
	}
	return &meta
}

type fetchDynamicListItemsParams struct {
	Ctx
	id string
}

func (c APIClient) fetchDynamicListItems(params *fetchDynamicListItemsParams) (APIResponse[FetchListItemsData], error) {
	meta := GetDynamicListMeta(params.id)
	if meta == nil {
		return newAPIResponse(nil, FetchListItemsData{}), errors.New("invalid id")
	}

	items := FetchListItemsData{}

	path := meta.Endpoint
	if meta.HasPeriod {
		path = strings.Replace(path, "{period}", meta.Period, 1)
	}
	if meta.HasUserId {
		path = strings.Replace(path, "{user_id}", meta.UserId, 1)
	}

	hasMore := true
	limit := 100
	page := 1
	maxPage := 5
	var res *http.Response
	var err error
	for hasMore {
		log.Debug("fetching dynamic list page", "id", params.id, "page", page)

		p := Ctx{}
		p.Query = &url.Values{}
		p.Query.Set("extended", "full,images")
		if !meta.NoPage {
			p.Query.Set("page", strconv.Itoa(page))
		}
		if !meta.NoLimit {
			p.Query.Set("limit", strconv.Itoa(limit))
		}

		switch meta.Endpoint {
		case dynamicListMetaById["movies/popular"].Endpoint, dynamicListMetaById["movies/recommendations"].Endpoint:
			response := listResponseData[ListItemMovie]{}
			res, err = c.Request("GET", path, p, &response)
			if err != nil {
				break
			}

			for i := range response.data {
				item := ListItem{}
				item.Type = meta.ItemType
				item.Movie = &response.data[i]
				items = append(items, item)
			}

		case dynamicListMetaById["shows/popular"].Endpoint, dynamicListMetaById["shows/recommendations"].Endpoint:
			response := listResponseData[ListItemShow]{}
			res, err = c.Request("GET", path, p, &response)
			if err != nil {
				break
			}

			for i := range response.data {
				item := ListItem{}
				item.Type = meta.ItemType
				item.Show = &response.data[i]
				items = append(items, item)
			}

		default:
			response := listResponseData[ListItem]{}
			res, err = c.Request("GET", path, p, &response)
			if err != nil {
				break
			}

			for i := range response.data {
				item := &response.data[i]
				if meta.ItemType != "" {
					item.Type = meta.ItemType
				}
				items = append(items, *item)
			}
		}

		hasMore = !meta.NoPage && page < maxPage && res.Header.Get("X-Pagination-Page") != res.Header.Get("X-Pagination-Page-Count")
		page++
	}

	return newAPIResponse(res, items), err
}
