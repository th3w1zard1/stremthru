package stremio_list

import (
	"bytes"
	"html/template"
	"net/http"
	"strconv"

	"github.com/MunifTanjim/stremthru/internal/anilist"
	"github.com/MunifTanjim/stremthru/internal/config"
	"github.com/MunifTanjim/stremthru/internal/mdblist"
	"github.com/MunifTanjim/stremthru/internal/oauth"
	"github.com/MunifTanjim/stremthru/internal/stremio/configure"
	stremio_shared "github.com/MunifTanjim/stremthru/internal/stremio/shared"
	stremio_template "github.com/MunifTanjim/stremthru/internal/stremio/template"
	stremio_userdata "github.com/MunifTanjim/stremthru/internal/stremio/userdata"
	"github.com/MunifTanjim/stremthru/internal/trakt"
	"github.com/google/uuid"
)

var IsPublicInstance = config.IsPublicInstance
var MaxPublicInstanceListCount = 10
var TraktEnabled = config.Integration.Trakt.IsEnabled()
var AniListEnabled = config.Feature.IsEnabled("anime")

type Base = stremio_template.BaseData

type TemplateDataList struct {
	Id      string
	URL     string
	Name    string
	Shuffle configure.Config
	Error   struct {
		URL  string
		Name string
	}
	Disabled struct {
		URL bool
	}
}

func newTemplateDataList(index int) TemplateDataList {
	return TemplateDataList{
		Shuffle: configure.Config{
			Key:   "lists[" + strconv.Itoa(index) + "].shuffle",
			Type:  configure.ConfigTypeCheckbox,
			Title: "Shuffle Items",
		},
	}
}

type supportedServiceUrl struct {
	Pattern  string
	Examples []string
}

type supportedService struct {
	Name     string
	Hostname string
	Icon     string
	URLs     []supportedServiceUrl
}

type TemplateData struct {
	Base

	Lists         []TemplateDataList
	CanAddList    bool
	CanRemoveList bool

	MDBListAPIKey configure.Config

	RPDBAPIKey configure.Config

	TraktEnabled bool
	TraktTokenId configure.Config

	Shuffle configure.Config

	ManifestURL string
	Script      template.JS

	CanAuthorize bool
	IsAuthed     bool
	AuthError    string

	SupportedServices []supportedService

	stremio_userdata.TemplateDataUserData
}

func (td *TemplateData) HasListError() bool {
	if len(td.Lists) == 0 {
		return true
	}
	for i := range td.Lists {
		if td.Lists[i].Error.URL != "" {
			return true
		}
	}
	if td.MDBListAPIKey.Error != "" {
		return true
	}
	return false
}

func (td *TemplateData) HasFieldError() bool {
	if td.HasListError() {
		return true
	}
	return false
}

func getTemplateData(ud *UserData, udError userDataError, isAuthed bool, r *http.Request) *TemplateData {
	td := &TemplateData{
		Base: Base{
			Title:       "StremThru List",
			Description: "Stremio Addon to access various Lists",
			NavTitle:    "List",
		},
		Lists: []TemplateDataList{},
		MDBListAPIKey: configure.Config{
			Key:          "mdblist_api_key",
			Type:         "password",
			Default:      ud.MDBListAPIkey,
			Title:        "API Key",
			Description:  `MDBList <a href="https://mdblist.com/preferences/#api_key_uid" target="_blank">API Key</a>`,
			Autocomplete: "off",
			Error:        udError.mdblist.api_key,
		},
		RPDBAPIKey: configure.Config{
			Key:          "rpdb_api_key",
			Type:         configure.ConfigTypePassword,
			Default:      ud.RPDBAPIKey,
			Title:        "API Key",
			Description:  `Rating Poster Database <a href="https://ratingposterdb.com/api-key/" target="blank">API Key</a>`,
			Autocomplete: "off",
		},
		TraktEnabled: TraktEnabled,
		TraktTokenId: configure.Config{
			Key:          "trakt_token_id",
			Title:        "Auth Code",
			Type:         configure.ConfigTypePassword,
			Default:      ud.TraktTokenId,
			Error:        udError.trakt_token_id,
			Autocomplete: "off",
			Action: configure.ConfigAction{
				Visible: ud.TraktTokenId == "" || udError.trakt_token_id != "",
				Label:   "Authorize",
				OnClick: template.JS(`window.open("` + oauth.TraktOAuthConfig.AuthCodeURL(uuid.NewString()) + `", "_blank")`),
			},
		},
		Shuffle: configure.Config{
			Key:   "shuffle",
			Type:  configure.ConfigTypeCheckbox,
			Title: "Shuffle Items for All Lists",
		},
		Script: ``,
	}

	if TraktEnabled && td.TraktTokenId.Error == "" {
		otok, err := ud.getTraktToken()
		if err != nil {
			td.TraktTokenId.Error = err.Error()
			td.TraktTokenId.Action.Visible = true
		} else if otok != nil {
			td.TraktTokenId.Title += " (" + otok.UserName + ")"
		}
	}

	if ud.Shuffle {
		td.Shuffle.Default = "checked"
	}

	hasListNames := len(ud.ListNames) > 0
	hasListShuffle := len(ud.ListShuffle) > 0
	for i, listId := range ud.Lists {
		list := newTemplateDataList(i)
		list.Id = listId
		if hasListNames {
			list.Name = ud.ListNames[i]
		}
		if hasListShuffle && ud.ListShuffle[i] == 1 {
			list.Shuffle.Default = "checked"
		}
		if len(ud.list_urls) > i {
			list.URL = ud.list_urls[i]
		}
		if len(udError.list_urls) > i {
			list.Error.URL = udError.list_urls[i]
		}

		if listId == "" {
			if list.Error.URL == "" {
				list.Error.URL = "Missing List ID"
			}
		} else if list.URL == "" {
			service, id, err := parseListId(listId)
			if err != nil {
				list.Error.URL = "Failed to Parse List ID: " + listId
			} else {
				switch service {
				case "anilist":
					l := anilist.AniListList{Id: id}
					if err := ud.FetchAniListList(&l, false); err != nil {
						log.Error("failed to fetch list", "error", err, "id", listId)
						list.Error.URL = "Failed to Fetch List: " + err.Error()
					} else {
						list.URL = l.GetURL()
					}
				case "mdblist":
					l := mdblist.MDBListList{Id: id}
					if err := ud.FetchMDBListList(&l); err != nil {
						log.Error("failed to fetch list", "error", err, "id", listId)
						list.Error.URL = "Failed to Fetch List: " + err.Error()
					} else {
						list.URL = l.GetURL()
					}

				case "trakt":
					if td.TraktTokenId.Error == "" {
						l := trakt.TraktList{Id: id}
						if err := ud.FetchTraktList(&l); err != nil {
							log.Error("failed to fetch list", "error", err, "id", listId)
							list.Error.URL = "Failed to Fetch List: " + err.Error()
						} else {
							list.URL = l.GetURL()
						}
					} else {
						list.Disabled.URL = true
						list.Error.URL = "Trakt.tv authorization needed"
					}
				}
			}
		}
		if list.URL == "" && list.Error.URL == "" {
			list.Error.URL = "Missing List URL"
		}
		td.Lists = append(td.Lists, list)
	}

	td.IsAuthed = isAuthed

	if udManager.IsSaved(ud) {
		td.SavedUserDataKey = udManager.GetId(ud)
	}
	if td.IsAuthed {
		if options, err := stremio_userdata.GetOptions("list"); err != nil {
			LogError(r, "failed to list saved userdata options", err)
		} else {
			td.SavedUserDataOptions = options
		}
	} else if td.SavedUserDataKey != "" {
		if sud, err := stremio_userdata.Get[UserData]("list", td.SavedUserDataKey); err != nil {
			LogError(r, "failed to get saved userdata", err)
		} else {
			td.SavedUserDataOptions = []configure.ConfigOption{{Label: sud.Name, Value: td.SavedUserDataKey}}
		}
	}

	return td
}

var executeTemplate = func() stremio_template.Executor[TemplateData] {
	return stremio_template.GetExecutor("stremio/list", func(td *TemplateData) *TemplateData {
		td.StremThruAddons = stremio_shared.GetStremThruAddons()
		td.Version = config.Version
		td.CanAuthorize = !IsPublicInstance
		td.CanAddList = td.IsAuthed || len(td.Lists) < MaxPublicInstanceListCount
		td.CanRemoveList = len(td.Lists) > 1

		td.SupportedServices = []supportedService{}
		if AniListEnabled {
			td.SupportedServices = append(td.SupportedServices, supportedService{
				Name:     "AniList",
				Hostname: "anilist.co",
				Icon:     "https://anilist.co/img/icons/favicon-32x32.png",
				URLs: []supportedServiceUrl{
					{
						Pattern:  "/user/{user_name}/animelist/{list_name}",
						Examples: []string{"/user/Kyokino/animelist/Films"},
					},
					{Pattern: "/search/anime/trending"},
					{Pattern: "/search/anime/this-season"},
					{Pattern: "/search/anime/next-season"},
					{Pattern: "/search/anime/popular"},
					{Pattern: "/search/anime/top-100"},
				},
			})
		}
		td.SupportedServices = append(td.SupportedServices, supportedService{
			Name:     "MDBList",
			Hostname: "mdblist.com",
			Icon:     "https://mdblist.com/static/favicon-32x32.png",
			URLs: []supportedServiceUrl{
				{
					Pattern: "/?list={list_id}",
					Examples: []string{
						"/?list=14",
					},
				},
				{
					Pattern: "/lists/{user_name}/{list_slug}",
					Examples: []string{
						"/lists/garycrawfordgc/latest-tv-shows",
					},
				},
				{
					Pattern: "/watchlist/{user_name}",
					Examples: []string{
						"/watchlist/garycrawfordgc",
					},
				},
			},
		})
		if TraktEnabled {
			td.SupportedServices = append(td.SupportedServices, supportedService{
				Name:     "Trakt.tv",
				Hostname: "trakt.tv",
				Icon:     "https://walter-r2.trakt.tv/hotlink-ok/public/2024/favicon.png",
				URLs: []supportedServiceUrl{
					{
						Pattern: "/users/{user_slug}/favorites",
						Examples: []string{
							"/users/sean/favorites",
						},
					},
					{
						Pattern: "/users/{user_slug}/watchlist",
						Examples: []string{
							"/users/garycrawfordgc/watchlist",
						},
					},
					{
						Pattern: "/users/{user_slug}/lists/{list_slug}",
						Examples: []string{
							"/users/garycrawfordgc/lists/latest-releases",
						},
					},
					{Pattern: "/movies/boxoffice"},
					{
						Pattern: "/{movies,shows}/anticipated",
						Examples: []string{
							"/movies/anticipated",
							"/shows/anticipated",
						},
					},
					{
						Pattern: "/{movies,shows}/collected/{period}",
						Examples: []string{
							"/movies/collected/daily",
							"/shows/collected/weekly",
						},
					},
					{
						Pattern: "/{movies,shows}/favorited/{period}",
						Examples: []string{
							"/movies/favorited/weekly",
							"/shows/favorited/monthly",
						},
					},
					{
						Pattern: "/{movies,shows}/popular",
						Examples: []string{
							"/movies/popular",
							"/shows/popular",
						},
					},
					{
						Pattern: "/{movies,shows}/recommendations",
						Examples: []string{
							"/movies/recommendations",
							"/shows/recommendations",
						},
					},
					{
						Pattern: "/{movies,shows}/trending",
						Examples: []string{
							"/movies/trending",
							"/shows/trending",
						},
					},
					{
						Pattern: "/{movies,shows}/watched/{period}",
						Examples: []string{
							"/movies/watched/monthly",
							"/shows/watched/all",
						},
					},
				},
			})
		}

		if len(td.Lists) == 0 {
			td.Lists = append(td.Lists, newTemplateDataList(0))
		}

		td.IsRedacted = !td.IsAuthed && td.SavedUserDataKey != ""
		if td.IsRedacted {
			redacted := "*******"
			if td.MDBListAPIKey.Default != "" {
				td.MDBListAPIKey.Default = redacted
			}
			if td.RPDBAPIKey.Default != "" {
				td.RPDBAPIKey.Default = redacted
			}
		}

		return td
	}, template.FuncMap{}, "configure_config.html", "configure_submit_button.html", "saved_userdata_field.html", "list.html")
}()

func getPage(td *TemplateData) (bytes.Buffer, error) {
	return executeTemplate(td, "list.html")
}
