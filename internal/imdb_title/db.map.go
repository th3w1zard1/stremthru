package imdb_title

import (
	"fmt"

	"github.com/MunifTanjim/stremthru/internal/db"
)

const MapTableName = "imdb_title_map"

type IMDBTitleMap struct {
	IMDBId    string       `json:"imdb"`
	TMDBId    string       `json:"tmdb"`
	TVDBId    string       `json:"tvdb"`
	TraktId   string       `json:"trakt"`
	UpdatedAt db.Timestamp `json:"uat"`
}

type MapColumnStruct struct {
	IMDBId    string
	TMDBId    string
	TVDBId    string
	TraktId   string
	UpdatedAt string
}

var MapColumn = MapColumnStruct{
	IMDBId:    "imdb",
	TMDBId:    "tmdb",
	TVDBId:    "tvdb",
	TraktId:   "trakt",
	UpdatedAt: "uat",
}

func RecordMappingFromMDBList(tx *db.Tx, imdbId string, tmdbId string, tvdbId string) error {
	query := fmt.Sprintf(
		`INSERT INTO %s (%s, %s, %s) VALUES (?, ?, ?) ON CONFLICT (%s) DO UPDATE SET %s = EXCLUDED.%s, %s = EXCLUDED.%s, %s = %s`,
		MapTableName,
		MapColumn.IMDBId,
		MapColumn.TMDBId,
		MapColumn.TVDBId,
		MapColumn.IMDBId,
		MapColumn.TMDBId,
		MapColumn.TMDBId,
		MapColumn.TVDBId,
		MapColumn.TVDBId,
		MapColumn.UpdatedAt,
		db.CurrentTimestamp,
	)

	_, err := tx.Exec(query, imdbId, tmdbId, tvdbId)
	return err
}
