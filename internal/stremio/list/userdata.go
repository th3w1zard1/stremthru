package stremio_list

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/MunifTanjim/stremthru/internal/anilist"
	"github.com/MunifTanjim/stremthru/internal/config"
	"github.com/MunifTanjim/stremthru/internal/mdblist"
	stremio_userdata "github.com/MunifTanjim/stremthru/internal/stremio/userdata"
)

type UserData struct {
	Lists        []string `json:"lists"`
	ListNames    []string `json:"list_names"`
	list_urls    []string `json:"-"`
	MDBListLists []int    `json:"mdblist_lists,omitempty"` // deprecated

	MDBListAPIkey string `json:"mdblist_api_key"`
	RPDBAPIKey    string `json:"rpdb_api_key,omitempty"`

	Shuffle bool `json:"shuffle,omitempty"`

	encoded string `json:"-"` // correctly configured

	mdblistById map[int]mdblist.MDBListList    `json:"-"`
	anilistById map[string]anilist.AniListList `json:"-"`
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
	list_urls []string
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

		ud.MDBListAPIkey = r.Form.Get("mdblist_api_key")

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
		isAniListEnabled := config.Feature.IsEnabled("anime")

		if isMDBListEnabled {
			userParams := mdblist.GetMyLimitsParams{}
			userParams.APIKey = ud.MDBListAPIkey
			if _, userErr := mdblistClient.GetMyLimits(&userParams); userErr != nil {
				err := userDataError{}
				err.mdblist.api_key = "Invalid API Key: " + userErr.Error()
				return ud, err
			}
		}

		ud.Lists = make([]string, 0, lists_length)
		if isAuthed {
			ud.ListNames = make([]string, 0, lists_length)
		}

		ud.list_urls = make([]string, 0, lists_length)
		udErr := userDataError{}
		udErr.list_urls = make([]string, 0, lists_length)

		idx := -1
		for i := range lists_length {
			listUrlStr := r.Form.Get("lists[" + strconv.Itoa(i) + "].url")
			if listUrlStr == "" {
				continue
			}

			idx++

			ud.Lists = append(ud.Lists, "")
			if isAuthed {
				ud.ListNames = append(ud.ListNames, r.Form.Get("lists["+strconv.Itoa(i)+"].name"))
			}

			ud.list_urls = append(ud.list_urls, listUrlStr)
			udErr.list_urls = append(udErr.list_urls, "")

			listUrl, err := url.Parse(listUrlStr)
			if err != nil {
				udErr.list_urls[idx] = "Invalid List URL: " + err.Error()
				continue
			}

			switch listUrl.Hostname() {
			case "anilist.co":
				if !isAniListEnabled {
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
					id, err := strconv.Atoi(idStr)
					if err != nil {
						udErr.list_urls[idx] = "Invalid List ID: " + err.Error()
						continue
					}
					list.Id = id
				} else if strings.HasPrefix(listUrl.Path, "/lists/") {
					username, slug, _ := strings.Cut(strings.TrimPrefix(listUrl.Path, "/lists/"), "/")
					if username != "" && slug != "" && !strings.Contains(slug, "/") {
						list.UserName = username
						list.Slug = slug
					} else {
						udErr.list_urls[idx] = "Invalid List URL"
						continue
					}
				} else {
					udErr.list_urls[idx] = "Invalid List URL"
					continue
				}

				err := ud.FetchMDBListList(&list)
				if err != nil {
					udErr.list_urls[idx] = "Failed to fetch List: " + err.Error()
					continue
				}
				ud.Lists[idx] = "mdblist:" + strconv.Itoa(list.Id)
			default:
				udErr.list_urls[idx] = "Unsupported List URL"
			}
		}

		if udErr.HasError() {
			return ud, udErr
		}

		ud.RPDBAPIKey = r.Form.Get("rpdb_api_key")
		ud.Shuffle = r.Form.Get("shuffle") == "on"
	}

	if IsPublicInstance && len(ud.Lists) > MaxPublicInstanceListCount {
		ud.Lists = ud.Lists[0:MaxPublicInstanceListCount]
	}

	return ud, nil
}

func (ud *UserData) FetchMDBListList(list *mdblist.MDBListList) error {
	if ud.mdblistById == nil {
		ud.mdblistById = map[int]mdblist.MDBListList{}
	}
	if list.Id != 0 {
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
