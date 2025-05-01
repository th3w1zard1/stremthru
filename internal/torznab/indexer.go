package torznab

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/MunifTanjim/stremthru/internal/buddy"
	"github.com/MunifTanjim/stremthru/internal/db"
	"github.com/MunifTanjim/stremthru/internal/imdb_torrent"
	"github.com/MunifTanjim/stremthru/internal/torrent_info"
)

type Info struct {
	ID          string
	Title       string
	Description string
	Link        string
	Language    string
	Category    string
}

type Indexer interface {
	Info() Info
	Search(query Query) ([]ResultItem, error)
	Download(urlStr string) (io.ReadCloser, http.Header, error)
	Capabilities() Caps
}

type stremThruIndexer struct {
	info Info
	caps Caps
}

func (sti stremThruIndexer) Info() Info {
	return sti.info
}

func (sti stremThruIndexer) Search(q Query) ([]ResultItem, error) {
	buddy.PullTorrentsByStremId(q.IMDBId, "")

	args := []any{}
	var query strings.Builder
	query.WriteString(
		fmt.Sprintf(
			"SELECT %s FROM %s ito INNER JOIN %s ti ON ti.%s = ito.%s WHERE ito.%s = ?",
			db.JoinPrefixedColumnNames("ti.", torrent_info.Columns...),
			imdb_torrent.TableName,
			torrent_info.TableName,
			torrent_info.Column.Hash,
			imdb_torrent.Column.Hash,
			imdb_torrent.Column.TId,
		),
	)
	args = append(args, q.IMDBId)
	if q.Season != "" {
		query.WriteString(
			fmt.Sprintf(
				" AND (ti.%s = ? OR CONCAT(',', ti.%s, ',') LIKE ?)",
				torrent_info.Column.Seasons,
				torrent_info.Column.Seasons,
			),
		)
		args = append(args, q.Season, "%,"+q.Season+",%")
	}
	if q.Ep != "" {
		if q.Season != "" {
			query.WriteString(
				fmt.Sprintf(
					" AND (ti.%s = '' OR ti.%s = ? OR CONCAT(',', ti.%s, ',') LIKE ?)",
					torrent_info.Column.Episodes,
					torrent_info.Column.Episodes,
					torrent_info.Column.Episodes,
				),
			)
			args = append(args, q.Ep, "%,"+q.Ep+",%")
		} else {
			query.WriteString(
				fmt.Sprintf(
					" AND (ti.%s = ? OR CONCAT(',', ti.%s, ',') LIKE ?)",
					torrent_info.Column.Episodes,
					torrent_info.Column.Episodes,
				),
			)
			args = append(args, q.Ep, "%,"+q.Ep+",%")
		}
	}
	query.WriteString(
		fmt.Sprintf(
			" AND ti.%s != -1",
			torrent_info.Column.Size,
		),
	)
	rows, err := db.Query(query.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []ResultItem{}
	for rows.Next() {
		var tInfo torrent_info.TorrentInfo
		if err := rows.Scan(
			&tInfo.Hash,
			&tInfo.TorrentTitle,

			&tInfo.Source,
			&tInfo.Category,
			&tInfo.CreatedAt,
			&tInfo.UpdatedAt,
			&tInfo.ParsedAt,
			&tInfo.ParserVersion,
			&tInfo.ParserInput,

			&tInfo.Audio,
			&tInfo.BitDepth,
			&tInfo.Channels,
			&tInfo.Codec,
			&tInfo.Commentary,
			&tInfo.Complete,
			&tInfo.Container,
			&tInfo.Convert,
			&tInfo.Date,
			&tInfo.Documentary,
			&tInfo.Dubbed,
			&tInfo.Edition,
			&tInfo.EpisodeCode,
			&tInfo.Episodes,
			&tInfo.Extended,
			&tInfo.Extension,
			&tInfo.Group,
			&tInfo.HDR,
			&tInfo.Hardcoded,
			&tInfo.Languages,
			&tInfo.Network,
			&tInfo.Proper,
			&tInfo.Quality,
			&tInfo.Region,
			&tInfo.Remastered,
			&tInfo.Repack,
			&tInfo.Resolution,
			&tInfo.Retail,
			&tInfo.Seasons,
			&tInfo.Site,
			&tInfo.Size,
			&tInfo.Subbed,
			&tInfo.ThreeD,
			&tInfo.Title,
			&tInfo.Uncensored,
			&tInfo.Unrated,
			&tInfo.Upscaled,
			&tInfo.Volumes,
			&tInfo.Year,
			&tInfo.YearEnd,
		); err != nil {
			return nil, err
		}
		var category Category
		switch tInfo.Category {
		case torrent_info.TorrentInfoCategoryMovie:
			category = CategoryMovies
		case torrent_info.TorrentInfoCategorySeries:
			category = CategoryTV
		case torrent_info.TorrentInfoCategoryXXX:
			category = CategoryXXX
		default:
			category = CategoryOther
		}
		audio := strings.Join(tInfo.Audio, ", ")
		if len(tInfo.Channels) > 0 {
			audio += " | " + strings.Join(tInfo.Channels, ", ")
		}
		items = append(items, ResultItem{
			Audio:       audio,
			Category:    category,
			IMDB:        q.IMDBId,
			InfoHash:    tInfo.Hash,
			Language:    strings.Join(tInfo.Languages, ", "),
			PublishDate: tInfo.CreatedAt.Time,
			Resolution:  tInfo.Resolution,
			Site:        tInfo.Site,
			Size:        tInfo.Size,
			Title:       tInfo.TorrentTitle,
			Year:        tInfo.Year,
		})
	}

	return items, nil
}

func (sti stremThruIndexer) Download(urlStr string) (io.ReadCloser, http.Header, error) {
	return nil, nil, nil
}

func (sti stremThruIndexer) Capabilities() Caps {
	return sti.caps
}

var StremThruIndexer = stremThruIndexer{
	caps: Caps{
		Searching: []CapsSearchingItem{
			{
				Name:            "tv-search",
				Available:       true,
				SupportedParams: []string{"imdbid,season,ep"},
			},
			{
				Name:            "movie-search",
				Available:       true,
				SupportedParams: []string{"imdbid"},
			},
		},
		Categories: []CapsCategory{
			{
				Category: CategoryMovies,
				Sub: []Category{
					CategoryMovies_Foreign,
					CategoryMovies_Other,
					CategoryMovies_SD,
					CategoryMovies_HD,
					CategoryMovies_3D,
					CategoryMovies_BluRay,
					CategoryMovies_DVD,
					CategoryMovies_WEBDL,
				},
			},
			{
				Category: CategoryTV,
				Sub: []Category{
					CategoryTV_WEBDL,
					CategoryTV_FOREIGN,
					CategoryTV_SD,
					CategoryTV_HD,
					CategoryTV_Other,
					CategoryTV_Sport,
					CategoryTV_Anime,
					CategoryTV_Documentary,
				},
			},
		},
	},
}
