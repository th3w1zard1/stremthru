package anilist

import (
	"context"
	"slices"

	"github.com/MunifTanjim/stremthru/internal/config"
	"github.com/hasura/go-graphql-client"
)

var client = graphql.NewClient(
	"https://graphql.anilist.co/graphql",
	config.GetHTTPClient(config.TUNNEL_TYPE_AUTO),
).WithDebug(true)

type getAnimeListQuery struct {
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

func FetchLists(userName string) ([]List, error) {
	var q getAnimeListQuery
	err := client.Query(context.Background(), &q, map[string]any{
		"userName": userName,
	})
	if err != nil {
		return nil, err
	}
	lists := make([]List, len(q.MediaListCollection.Lists))
	for i := range q.MediaListCollection.Lists {
		l := &q.MediaListCollection.Lists[i]
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
		lists[i] = list
	}
	return lists, nil
}

type fetchMediasQuery struct {
	Page struct {
		Media []struct {
			Id    int
			IdMal int
			Type  string
			Title struct {
				English string
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
			medias = append(medias, media)
		}
	}
	return medias, nil
}
