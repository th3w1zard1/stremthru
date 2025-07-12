package imdb_title

import (
	"fmt"
	"strings"

	"github.com/MunifTanjim/stremthru/internal/db"
	"github.com/MunifTanjim/stremthru/internal/util"
)

const MapTableName = "imdb_title_map"

type IMDBTitleMap struct {
	IMDBId    string       `json:"imdb"`
	TMDBId    string       `json:"tmdb"`
	TVDBId    string       `json:"tvdb"`
	TraktId   string       `json:"trakt"`
	MALId     string       `json:"mal"`
	UpdatedAt db.Timestamp `json:"uat"`
}

type MapColumnStruct struct {
	IMDBId    string
	TMDBId    string
	TVDBId    string
	TraktId   string
	MALId     string
	UpdatedAt string
}

var MapColumn = MapColumnStruct{
	IMDBId:    "imdb",
	TMDBId:    "tmdb",
	TVDBId:    "tvdb",
	TraktId:   "trakt",
	MALId:     "mal",
	UpdatedAt: "uat",
}

var query_get_imdb_id_by_trakt_id = fmt.Sprintf(
	`SELECT %s, %s FROM %s WHERE %s IN `,
	MapColumn.IMDBId,
	MapColumn.TraktId,
	MapTableName,
	MapColumn.TraktId,
)

func GetIMDBIdByTraktId(traktIds []string) (map[string]string, error) {
	count := len(traktIds)
	if count == 0 {
		return nil, nil
	}

	query := query_get_imdb_id_by_trakt_id + "(" + util.RepeatJoin("?", count, ",") + ")"
	args := make([]any, count)
	for i := range traktIds {
		args[i] = traktIds[i]
	}
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	imdbIdByTraktId := make(map[string]string, count)
	for rows.Next() {
		var imdbId string
		var traktId string
		if err := rows.Scan(&imdbId, &traktId); err != nil {
			return nil, err
		}
		imdbIdByTraktId[traktId] = imdbId
	}

	return imdbIdByTraktId, nil
}

func RecordMappingFromMDBList(tx *db.Tx, imdbId, tmdbId, tvdbId, traktId, malId string) error {
	query := fmt.Sprintf(
		`INSERT INTO %s AS itm (%s) VALUES (?,?,?,?,?) ON CONFLICT (%s) DO UPDATE SET %s, %s = %s`,
		MapTableName,
		db.JoinColumnNames(MapColumn.IMDBId, MapColumn.TMDBId, MapColumn.TVDBId, MapColumn.TraktId, MapColumn.MALId),
		MapColumn.IMDBId,
		strings.Join(
			[]string{
				fmt.Sprintf("%s = CASE WHEN itm.%s = '' THEN EXCLUDED.%s ELSE itm.%s END", MapColumn.TMDBId, MapColumn.TMDBId, MapColumn.TMDBId, MapColumn.TMDBId),
				fmt.Sprintf("%s = CASE WHEN itm.%s = '' THEN EXCLUDED.%s ELSE itm.%s END", MapColumn.TVDBId, MapColumn.TVDBId, MapColumn.TVDBId, MapColumn.TVDBId),
				fmt.Sprintf("%s = CASE WHEN itm.%s = '' THEN EXCLUDED.%s ELSE itm.%s END", MapColumn.TraktId, MapColumn.TraktId, MapColumn.TraktId, MapColumn.TraktId),
				fmt.Sprintf("%s = CASE WHEN itm.%s = '' THEN EXCLUDED.%s ELSE itm.%s END", MapColumn.MALId, MapColumn.MALId, MapColumn.MALId, MapColumn.MALId),
			},
			", ",
		),
		MapColumn.UpdatedAt,
		db.CurrentTimestamp,
	)

	_, err := tx.Exec(query, imdbId, tmdbId, tvdbId, traktId, malId)
	return err
}

type BulkRecordMappingInputItem struct {
	IMDBId  string
	TMDBId  string
	TVDBId  string
	TraktId string
	MALId   string
}

var query_bulk_record_mapping_before_values = fmt.Sprintf(
	`INSERT INTO %s AS itm (%s,%s,%s,%s,%s) VALUES `,
	MapTableName,
	MapColumn.IMDBId,
	MapColumn.TMDBId,
	MapColumn.TVDBId,
	MapColumn.TraktId,
	MapColumn.MALId,
)
var query_bulk_record_mapping_placeholder = fmt.Sprintf(
	`(?,?,?,?,?)`,
)
var query_bulk_record_mapping_after_values = fmt.Sprintf(
	` ON CONFLICT (%s) DO UPDATE SET %s, %s = %s`,
	MapColumn.IMDBId,
	strings.Join(
		[]string{
			fmt.Sprintf("%s = CASE WHEN itm.%s = '' THEN EXCLUDED.%s ELSE itm.%s END", MapColumn.TMDBId, MapColumn.TMDBId, MapColumn.TMDBId, MapColumn.TMDBId),
			fmt.Sprintf("%s = CASE WHEN itm.%s = '' THEN EXCLUDED.%s ELSE itm.%s END", MapColumn.TVDBId, MapColumn.TVDBId, MapColumn.TVDBId, MapColumn.TVDBId),
			fmt.Sprintf("%s = CASE WHEN itm.%s = '' THEN EXCLUDED.%s ELSE itm.%s END", MapColumn.TraktId, MapColumn.TraktId, MapColumn.TraktId, MapColumn.TraktId),
			fmt.Sprintf("%s = CASE WHEN itm.%s = '' THEN EXCLUDED.%s ELSE itm.%s END", MapColumn.MALId, MapColumn.MALId, MapColumn.MALId, MapColumn.MALId),
		},
		", ",
	),
	MapColumn.UpdatedAt,
	db.CurrentTimestamp,
)

func normalizeOptionalId(id string) string {
	if id == "0" {
		return ""
	}
	return id
}

func BulkRecordMapping(items []BulkRecordMappingInputItem) {
	count := len(items)
	query := query_bulk_record_mapping_before_values +
		util.RepeatJoin(query_bulk_record_mapping_placeholder, count, ",") +
		query_bulk_record_mapping_after_values

	args := make([]any, count*5)
	for i, item := range items {
		args[i*5+0] = item.IMDBId
		args[i*5+1] = normalizeOptionalId(item.TMDBId)
		args[i*5+2] = normalizeOptionalId(item.TVDBId)
		args[i*5+3] = normalizeOptionalId(item.TraktId)
		args[i*5+4] = normalizeOptionalId(item.MALId)
	}

	_, err := db.Exec(query, args...)
	if err != nil {
		log.Error("failed to bulk record mapping", "error", err)
	}
}
