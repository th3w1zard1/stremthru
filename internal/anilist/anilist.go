package anilist

import (
	"context"
	"net/http"
	"slices"
	"time"

	"github.com/MunifTanjim/stremthru/internal/config"
	"github.com/hasura/go-graphql-client"
)

var client = graphql.NewClient(
	"https://graphql.anilist.co/graphql",
	config.GetHTTPClient(config.TUNNEL_TYPE_AUTO),
	graphql.WithRetry(3),
	graphql.WithRetryBaseDelay(2*time.Second),
	graphql.WithRetryExponentialRate(2),
	graphql.WithRetryHTTPStatus([]int{http.StatusTooManyRequests}),
).WithDebug(config.Environment == config.EnvDev)

type MediaSeason string

const (
	MediaSeasonWinter MediaSeason = "WINTER"
	MediaSeasonSpring MediaSeason = "SPRING"
	MediaSeasonSummer MediaSeason = "SUMMER"
	MediaSeasonFall   MediaSeason = "FALL"
)

type MediaSort string

const (
	MediaSortTrendingDesc   MediaSort = "TRENDING_DESC"
	MediaSortPopularityDesc MediaSort = "POPULARITY_DESC"
	MediaSortScoreDesc      MediaSort = "SCORE_DESC"
)

type ListMedia struct {
	Id    int
	Score int
}

type List struct {
	UserName       string
	Name           string
	IsCustom       bool
	MediaIds       []int
	ScoreByMediaId map[int]int
}

func (l List) GetId() string {
	return l.UserName + ":" + l.Name
}

type getUserAnimeListQuery struct {
	MediaListCollection struct {
		Lists []struct {
			Name         string
			IsCustomList bool
			Entries      []struct {
				Score     int
				MediaList struct {
					Media struct {
						Id int
					}
				} `graphql:"... on MediaList"`
			}
		}
	} `graphql:"MediaListCollection(userName: $userName, type: ANIME)"`
}

func FetchUserList(userName, name string) (*List, error) {
	var q getUserAnimeListQuery
	err := client.Query(context.Background(), &q, map[string]any{
		"userName": userName,
	})
	if err != nil {
		return nil, err
	}
	for i := range q.MediaListCollection.Lists {
		l := &q.MediaListCollection.Lists[i]
		if l.Name != name {
			continue
		}
		list := List{
			UserName:       userName,
			Name:           l.Name,
			IsCustom:       l.IsCustomList,
			MediaIds:       make([]int, len(l.Entries)),
			ScoreByMediaId: make(map[int]int, len(l.Entries)),
		}
		for i := range l.Entries {
			mediaId := l.Entries[i].MediaList.Media.Id
			list.MediaIds[i] = mediaId
			list.ScoreByMediaId[mediaId] = l.Entries[i].Score
		}
		return &list, nil
	}
	return nil, nil
}

const searchAnimeListMaxPage = 4
const searchAnimeListPerPage = 50
const searchAnimeListQuery = `query (
  $page: Int!
  $season: MediaSeason
  $seasonYear: Int
  $sort: [MediaSort]
) {
  Page(page: $page, perPage: 50) {
		media(type: ANIME, season: $season, seasonYear: $seasonYear, sort: $sort) {
      id
    }
  }
}`

type SearchAnimeListData struct {
	Page struct {
		Media []struct {
			Id int `json:"id"`
		} `json:"media"`
	} `json:"Page"`
}

func getSeason(month time.Month) MediaSeason {
	switch month {
	case time.January, time.February, time.March:
		return MediaSeasonWinter
	case time.April, time.May, time.June:
		return MediaSeasonSpring
	case time.July, time.August, time.September:
		return MediaSeasonSummer
	case time.October, time.November, time.December:
		return MediaSeasonFall
	}
	panic("unreachable")
}

type searchListMeta struct {
	getInput func(page int) map[string]any
	name     string
}

var searchListQueryInputByName = map[string]searchListMeta{
	"trending": {
		name: "Trending",
		getInput: func(page int) map[string]any {
			return map[string]any{
				"page": page,
				"sort": []MediaSort{MediaSortTrendingDesc, MediaSortPopularityDesc},
			}
		},
	},
	"this-season": {
		name: "Popular This Season",
		getInput: func(page int) map[string]any {
			t := time.Now()
			return map[string]any{
				"page":       page,
				"season":     getSeason(t.Month()),
				"seasonYear": t.Year(),
				"sort":       []MediaSort{MediaSortPopularityDesc, MediaSortScoreDesc},
			}
		},
	},
	"next-season": {
		name: "Upcoming Next Season",
		getInput: func(page int) map[string]any {
			t := time.Now().AddDate(0, 3, 0)
			return map[string]any{
				"page":       page,
				"season":     getSeason(t.Month()),
				"seasonYear": t.Year(),
				"sort":       []MediaSort{MediaSortPopularityDesc, MediaSortScoreDesc},
			}
		},
	},
	"popular": {
		name: "All Time Popular",
		getInput: func(page int) map[string]any {
			return map[string]any{
				"page": page,
				"sort": []MediaSort{MediaSortPopularityDesc},
			}
		},
	},
	"top-100": {
		name: "Top 100",
		getInput: func(page int) map[string]any {
			return map[string]any{
				"page": page,
				"sort": []MediaSort{MediaSortScoreDesc},
			}
		},
	},
}

func IsValidSearchList(name string) bool {
	_, ok := searchListQueryInputByName[name]
	return ok
}

func FetchSearchList(name string) (*List, error) {
	meta, ok := searchListQueryInputByName[name]
	if !ok {
		return nil, nil
	}
	totalItems := searchAnimeListMaxPage * searchAnimeListPerPage
	list := List{
		UserName:       "~",
		Name:           name,
		MediaIds:       make([]int, 0, totalItems),
		ScoreByMediaId: make(map[int]int, totalItems),
	}
	for pageIdx := range searchAnimeListMaxPage {
		page := pageIdx + 1
		log.Debug("fetching search list page", "name", name, "page", page)
		var data SearchAnimeListData
		err := client.Exec(context.Background(), searchAnimeListQuery, &data, meta.getInput(page))
		if err != nil {
			return nil, err
		}
		medias := data.Page.Media
		for mIdx := range medias {
			mediaId := medias[mIdx].Id
			list.MediaIds = append(list.MediaIds, mediaId)
			list.ScoreByMediaId[mediaId] = totalItems - page*searchAnimeListPerPage + mIdx
		}
		if len(medias) < searchAnimeListPerPage {
			break
		}
	}
	return &list, nil
}

type fetchMediasQuery struct {
	Page struct {
		Media []struct {
			Id    int
			IdMal int
			Type  string
			Title struct {
				English string
				Romaji  string
			}
			Description string
			BannerImage string
			CoverImage  struct {
				ExtraLarge string
			}
			Duration  int
			IsAdult   bool
			Genres    []string
			StartDate struct {
				Year int
			}
		} `graphql:"media(id_in: $ids)"`
	} `graphql:"Page(page: $page, perPage: 50)"`
}

type Media struct {
	Id          int
	IdMal       int
	Type        string
	Title       string
	Description string
	BannerImage string
	CoverImage  string
	Duration    int
	IsAdult     bool
	Genres      []string
	StartYear   int
}

func FetchMedias(mediaIds []int) ([]Media, error) {
	if len(mediaIds) == 0 {
		return nil, nil
	}

	medias := []Media{}
	for cIds := range slices.Chunk(mediaIds, 50) {
		var q fetchMediasQuery
		err := client.Query(context.Background(), &q, map[string]any{
			"page": 1,
			"ids":  cIds,
		})
		if err != nil {
			return nil, err
		}
		for i := range q.Page.Media {
			m := &q.Page.Media[i]
			media := Media{
				Id:          m.Id,
				IdMal:       m.IdMal,
				Type:        m.Type,
				Title:       m.Title.English,
				Description: m.Description,
				BannerImage: m.BannerImage,
				CoverImage:  m.CoverImage.ExtraLarge,
				Duration:    m.Duration,
				IsAdult:     m.IsAdult,
				Genres:      m.Genres,
				StartYear:   m.StartDate.Year,
			}
			if media.Title == "" {
				media.Title = m.Title.Romaji
			}
			medias = append(medias, media)
		}
	}
	return medias, nil
}

type MediaFormat string

const (
	MediaFormatTV      MediaFormat = "TV"
	MediaFormatTVShort MediaFormat = "TV_SHORT"
	MediaFormatMovie   MediaFormat = "MOVIE"
	MediaFormatSpecial MediaFormat = "SPECIAL"
	MediaFormatOVA     MediaFormat = "OVA"
	MediaFormatONA     MediaFormat = "ONA"
	MediaFormatMusic   MediaFormat = "MUSIC"
	MediaFormatManga   MediaFormat = "MANGA"
	MediaFormatNovel   MediaFormat = "NOVEL"
	MediaFormatOneShot MediaFormat = "ONE_SHOT"
)

type fetchAnimeMediaFormatInfoQuery struct {
	Page struct {
		Media []struct {
			Id     int
			IdMal  int
			Format string
		} `graphql:"media(type: ANIME, id_in: $ids)"`
	} `graphql:"Page(page: $page, perPage: 50)"`
}

type MediaFormatInfo struct {
	Id     int
	IdMal  int
	Format MediaFormat
}

func FetchAnimeMediaFormatInfo(mediaIds []int) ([]MediaFormatInfo, error) {
	if len(mediaIds) == 0 {
		return nil, nil
	}

	infos := []MediaFormatInfo{}
	for cIds := range slices.Chunk(mediaIds, 50) {
		var q fetchAnimeMediaFormatInfoQuery
		err := client.Query(context.Background(), &q, map[string]any{
			"page": 1,
			"ids":  cIds,
		})
		if err != nil {
			return nil, err
		}
		for i := range q.Page.Media {
			m := &q.Page.Media[i]
			info := MediaFormatInfo{
				Id:     m.Id,
				IdMal:  m.IdMal,
				Format: MediaFormat(m.Format),
			}
			infos = append(infos, info)
		}
		time.Sleep(2 * time.Second)
	}
	return infos, nil
}
