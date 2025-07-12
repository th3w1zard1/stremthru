package stremio_list

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/MunifTanjim/stremthru/internal/anilist"
	"github.com/MunifTanjim/stremthru/internal/mdblist"
	"github.com/MunifTanjim/stremthru/internal/oauth"
	stremio_userdata "github.com/MunifTanjim/stremthru/internal/stremio/userdata"
	"github.com/MunifTanjim/stremthru/internal/trakt"
)

type UserData struct {
	Lists        []string `json:"lists"`
	ListNames    []string `json:"list_names"`
	ListShuffle  []int    `json:"list_shuffle"`
	list_urls    []string `json:"-"`
	MDBListLists []int    `json:"mdblist_lists,omitempty"` // deprecated

	MDBListAPIkey string `json:"mdblist_api_key,omitempty"`

	TraktTokenId string            `json:"trakt_token_id,omitempty"`
	traktToken   *oauth.OAuthToken `json:"-"`

	RPDBAPIKey string `json:"rpdb_api_key,omitempty"`

	Shuffle bool `json:"shuffle,omitempty"`

	encoded string `json:"-"` // correctly configured

	mdblistById map[string]mdblist.MDBListList `json:"-"`
	anilistById map[string]anilist.AniListList `json:"-"`
	traktById   map[string]trakt.TraktList     `json:"-"`
}

var udManager = stremio_userdata.NewManager[UserData](&stremio_userdata.ManagerConfig{
	AddonName: "list",
})

func (ud UserData) HasRequiredValues() bool {
	return len(ud.Lists) != 0
}

func (ud *UserData) GetEncoded() string {
	return ud.encoded
}

func (ud *UserData) SetEncoded(encoded string) {
	ud.encoded = encoded
}

func (ud *UserData) Ptr() *UserData {
	return ud
}

type userDataError struct {
	mdblist struct {
		api_key string
	}
	list_urls      []string
	trakt_token_id string
}

func (uderr userDataError) HasError() bool {
	if uderr.mdblist.api_key != "" {
		return true
	}
	for i := range uderr.list_urls {
		if uderr.list_urls[i] != "" {
			return true
		}
	}
	return false
}

func (uderr userDataError) Error() string {
	var str strings.Builder
	if uderr.mdblist.api_key != "" {
		str.WriteString("mdblist.api_key: " + uderr.mdblist.api_key + "\n")
	}
	for i, err := range uderr.list_urls {
		if err != "" {
			str.WriteString("mdblist.list[" + strconv.Itoa(i) + "].url: " + err + "\n")
		}
	}
	return str.String()
}

func getUserData(r *http.Request, isAuthed bool) (*UserData, error) {
	ud := &UserData{}
	ud.SetEncoded(r.PathValue("userData"))

	if IsMethod(r, http.MethodGet) || IsMethod(r, http.MethodHead) {
		if err := udManager.Resolve(ud); err != nil {
			return nil, err
		}
		if ud.encoded == "" {
			return ud, nil
		}

		if len(ud.MDBListLists) > 0 {
			for _, id := range ud.MDBListLists {
				ud.Lists = append(ud.Lists, "mdblist:"+strconv.Itoa(id))
			}

			ud.MDBListLists = nil

			if err := udManager.Sync(ud); err != nil {
				return nil, err
			}
		}
	}

	if IsMethod(r, http.MethodPost) {
		err := r.ParseForm()
		if err != nil {
			return nil, err
		}

		udErr := userDataError{}

		ud.MDBListAPIkey = r.Form.Get("mdblist_api_key")
		ud.TraktTokenId = r.Form.Get("trakt_token_id")

		ud.RPDBAPIKey = r.Form.Get("rpdb_api_key")
		ud.Shuffle = r.Form.Get("shuffle") == "on"

		lists_length := 0
		if v := r.Form.Get("lists_length"); v != "" {
			if lists_length, err = strconv.Atoi(v); err != nil {
				return nil, err
			}
		}

		if lists_length == 0 {
			err := userDataError{}
			err.list_urls = []string{"Missing List URL"}
			return ud, err
		}

		isMDBListEnabled := ud.MDBListAPIkey != ""
		isTraktTvConfigured := TraktEnabled && ud.TraktTokenId != ""

		if isMDBListEnabled {
			userParams := mdblist.GetMyLimitsParams{}
			userParams.APIKey = ud.MDBListAPIkey
			if _, userErr := mdblistClient.GetMyLimits(&userParams); userErr != nil {
				udErr.mdblist.api_key = "Invalid API Key: " + userErr.Error()
			}
		}

		if isTraktTvConfigured {
			ud.traktToken, err = ud.getTraktToken()
			if err != nil {
				udErr.trakt_token_id = err.Error()
			}
			isTraktTvConfigured = ud.TraktTokenId != ""
		}

		ud.Lists = make([]string, 0, lists_length)
		if isAuthed {
			ud.ListNames = make([]string, 0, lists_length)
		}
		ud.ListShuffle = make([]int, 0, lists_length)

		ud.list_urls = make([]string, 0, lists_length)
		udErr.list_urls = make([]string, 0, lists_length)

		idx := -1
		for i := range lists_length {
			listId := r.Form.Get("lists[" + strconv.Itoa(i) + "].id")
			listUrlStr := r.Form.Get("lists[" + strconv.Itoa(i) + "].url")
			if listId == "" && listUrlStr == "" {
				continue
			}

			idx++

			ud.Lists = append(ud.Lists, listId)
			if isAuthed {
				ud.ListNames = append(ud.ListNames, r.Form.Get("lists["+strconv.Itoa(i)+"].name"))
			}
			if r.Form.Get("lists["+strconv.Itoa(i)+"].shuffle") == "on" {
				ud.ListShuffle = append(ud.ListShuffle, 1)
			} else {
				ud.ListShuffle = append(ud.ListShuffle, 0)
			}

			ud.list_urls = append(ud.list_urls, listUrlStr)
			udErr.list_urls = append(udErr.list_urls, "")

			if listUrlStr == "" {
				continue
			}

			listUrl, err := url.Parse(listUrlStr)
			if err != nil {
				udErr.list_urls[idx] = "Invalid List URL: " + err.Error()
				continue
			}

			switch listUrl.Hostname() {
			case "anilist.co":
				if !AniListEnabled {
					udErr.list_urls[idx] = "Unsupported List URL"
					continue
				}

				list := anilist.AniListList{}
				if strings.HasPrefix(listUrl.Path, "/user/") {
					parts := strings.SplitN(strings.TrimPrefix(listUrl.Path, "/user/"), "/", 3)
					if len(parts) != 3 || parts[1] != "animelist" {
						udErr.list_urls[idx] = "Invalid AniList URL"
						continue
					}
					userName, listName := parts[0], parts[2]
					if userName == "" || listName == "" {
						udErr.list_urls[idx] = "Invalid AniList URL"
						continue
					}
					list.Id = userName + ":" + listName
				} else if strings.HasPrefix(listUrl.Path, "/search/anime/") {
					name := strings.TrimPrefix(listUrl.Path, "/search/anime/")
					if !anilist.IsValidSearchList(name) {
						udErr.list_urls[idx] = "Unsupported AniList URL"
						continue
					}
					list.Id = "~:" + name
				} else {
					udErr.list_urls[idx] = "Unsupported AniList URL"
					continue
				}

				err := ud.FetchAniListList(&list, true)
				if err != nil {
					udErr.list_urls[idx] = "Failed to fetch List: " + err.Error()
					continue
				}
				ud.Lists[idx] = "anilist:" + list.Id
			case "mdblist.com":
				if !isMDBListEnabled {
					udErr.list_urls[idx] = "MDBList API Key is required"
					continue
				}

				query := listUrl.Query()
				list := mdblist.MDBListList{}
				if idStr := query.Get("list"); idStr != "" {
					list.Id = idStr
				} else if strings.HasPrefix(listUrl.Path, "/lists/") {
					username, slug, _ := strings.Cut(strings.TrimPrefix(listUrl.Path, "/lists/"), "/")
					if username != "" && slug != "" && !strings.Contains(slug, "/") {
						list.UserName = username
						list.Slug = slug
					} else {
						udErr.list_urls[idx] = "Invalid List URL"
						continue
					}
				} else if strings.HasPrefix(listUrl.Path, "/watchlist/") {
					username := strings.TrimPrefix(listUrl.Path, "/watchlist/")
					list.Id = "~:watchlist:" + username
					list.UserName = username
					list.Slug = "watchlist/" + username
				} else {
					udErr.list_urls[idx] = "Invalid List URL"
					continue
				}

				err := ud.FetchMDBListList(&list)
				if err != nil {
					udErr.list_urls[idx] = "Failed to fetch List: " + err.Error()
					continue
				}
				ud.Lists[idx] = "mdblist:" + list.Id

			case "trakt.tv":
				if !isTraktTvConfigured {
					if TraktEnabled {
						udErr.list_urls[idx] = "Trakt.tv Auth Code is required"
					} else {
						udErr.list_urls[idx] = "Unsupported List URL"
					}
					continue
				}

				list := trakt.TraktList{}
				switch {
				case strings.HasPrefix(listUrl.Path, "/users/"):
					parts := strings.SplitN(strings.TrimPrefix(listUrl.Path, "/users/"), "/", 3)
					switch {
					case len(parts) == 3 && parts[1] == "lists":
						userSlug, listSlug := parts[0], parts[2]
						if userSlug == "" || listSlug == "" {
							udErr.list_urls[idx] = "Invalid Trakt.tv URL"
							continue
						}
						list.UserId = userSlug
						list.Slug = listSlug

					case len(parts) == 2:
						switch parts[1] {
						case "collection", "favorites", "watchlist":
							list.Id = "~:" + parts[1] + ":" + parts[0]
							list.UserId = parts[0]
						default:
							udErr.list_urls[idx] = "Unsupported Trakt.tv URL"
							continue
						}
					default:
						udErr.list_urls[idx] = "Unsupported Trakt.tv URL"
						continue
					}

				default:
					meta := trakt.GetDynamicListMeta(listUrl.Path)
					if meta == nil {
						udErr.list_urls[idx] = "Unsupported Trakt.tv URL"
						continue
					}

					list.Id = meta.Id
					if list.Id == "" {
						list.Id = "~:" + strings.TrimPrefix(listUrl.Path, "/")
					}
				}

				err := ud.FetchTraktList(&list)
				if err != nil {
					udErr.list_urls[idx] = "Failed to fetch List: " + err.Error()
					continue
				}
				ud.Lists[idx] = "trakt:" + list.Id
			}
		}

		if udErr.HasError() {
			return ud, udErr
		}
	}

	if IsPublicInstance && len(ud.Lists) > MaxPublicInstanceListCount {
		ud.Lists = ud.Lists[0:MaxPublicInstanceListCount]
	}

	return ud, nil
}

func (ud *UserData) getTraktToken() (*oauth.OAuthToken, error) {
	if ud.TraktTokenId == "" {
		return nil, nil
	}

	if ud.traktToken != nil {
		return ud.traktToken, nil
	}

	otok, err := oauth.GetOAuthTokenById(ud.TraktTokenId)
	if err != nil {
		ud.TraktTokenId = ""
		return nil, errors.New("failed to retrieve token: " + err.Error())
	} else if otok != nil && otok.IsExpired() {
		traktClient := trakt.GetAPIClient(otok.Id)
		settings, err := traktClient.RetrieveSettings(&trakt.RetrieveSettingsParams{})
		if err != nil || settings.Data.User.Ids.Slug != otok.UserId {
			otok.AccessToken = ""
			otok.RefreshToken = ""
			err = oauth.SaveOAuthToken(otok)
			if err != nil {
				log.Error("failed to delete trakt token", "error", err, "id", otok.Id)
			}
			otok = nil
		}
	}
	if otok == nil {
		ud.TraktTokenId = ""
		return nil, errors.New("Invalid or Revoked")
	}

	ud.traktToken = otok
	return ud.traktToken, nil
}

func (ud *UserData) FetchMDBListList(list *mdblist.MDBListList) error {
	if ud.mdblistById == nil {
		ud.mdblistById = map[string]mdblist.MDBListList{}
	}
	if list.Id != "" {
		if l, ok := ud.mdblistById[list.Id]; ok {
			*list = l
			return nil
		}
	}
	if err := list.Fetch(ud.MDBListAPIkey); err != nil {
		return err
	}
	ud.mdblistById[list.Id] = *list
	return nil
}

func (ud *UserData) FetchAniListList(list *anilist.AniListList, scheduleIdMapSync bool) error {
	if ud.anilistById == nil {
		ud.anilistById = map[string]anilist.AniListList{}
	}
	if list.Id != "" {
		if l, ok := ud.anilistById[list.Id]; ok {
			*list = l
			return nil
		}
	}
	if err := list.Fetch(); err != nil {
		return err
	}

	if scheduleIdMapSync {
		anilist.ScheduleIdMapSync(list.Medias)
	}

	ud.anilistById[list.Id] = *list
	return nil
}

func (ud *UserData) FetchTraktList(list *trakt.TraktList) error {
	if ud.traktById == nil {
		ud.traktById = map[string]trakt.TraktList{}
	}
	if list.Id != "" {
		if l, ok := ud.traktById[list.Id]; ok {
			*list = l
			return nil
		}
	}
	if err := list.Fetch(ud.TraktTokenId); err != nil {
		return err
	}

	ud.traktById[list.Id] = *list
	return nil
}
