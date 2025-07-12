package anime

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/MunifTanjim/stremthru/internal/anidb"
	"github.com/MunifTanjim/stremthru/internal/db"
	"github.com/MunifTanjim/stremthru/internal/util"
)

const IdMapTableName = "anime_id_map"

type AnimeIdMapType = string

const (
	AnimeIdMapTypeTV      AnimeIdMapType = "TV"
	AnimeIdMapTypeTVShort AnimeIdMapType = "TV_SHORT"
	AnimeIdMapTypeMovie   AnimeIdMapType = "MOVIE"
	AnimeIdMapTypeSpecial AnimeIdMapType = "SPECIAL"
	AnimeIdMapTypeOVA     AnimeIdMapType = "OVA"
	AnimeIdMapTypeONA     AnimeIdMapType = "ONA"
	AnimeIdMapTypeMusic   AnimeIdMapType = "MUSIC"
	AnimeIdMapTypeManga   AnimeIdMapType = "MANGA"
	AnimeIdMapTypeNovel   AnimeIdMapType = "NOVEL"
	AnimeIdMapTypeOneShot AnimeIdMapType = "ONE_SHOT"
	AnimeIdMapTypeUnknown AnimeIdMapType = ""
)

type AnimeIdMap struct {
	Id          int            `json:"id"`
	Type        AnimeIdMapType `json:"type"`
	AniList     string         `json:"anilist"`
	AniDB       string         `json:"anidb"`
	AniSearch   string         `json:"anisearch"`
	AnimePlanet string         `json:"animeplanet"`
	IMDB        string         `json:"imdb"`
	Kitsu       string         `json:"kitsu"`
	LiveChart   string         `json:"livechart"`
	MAL         string         `json:"mal"`
	NotifyMoe   string         `json:"notifymoe"`
	TMDB        string         `json:"tmdb"`
	TVDB        string         `json:"tvdb"`
	UpdatedAt   db.Timestamp   `json:"uat"`
}

func (idMap *AnimeIdMap) IsZero() bool {
	return idMap.Id == 0
}

func (idMap *AnimeIdMap) IsStale() bool {
	return time.Now().After(idMap.UpdatedAt.Add(15 * 24 * time.Hour))
}

type rawAnimeIdMap struct {
	Id          int            `json:"id"`
	Type        AnimeIdMapType `json:"type"`
	AniList     db.NullString  `json:"anilist"`
	AniDB       db.NullString  `json:"anidb"`
	AniSearch   db.NullString  `json:"anisearch"`
	AnimePlanet db.NullString  `json:"animeplanet"`
	IMDB        db.NullString  `json:"imdb"`
	Kitsu       db.NullString  `json:"kitsu"`
	LiveChart   db.NullString  `json:"livechart"`
	MAL         db.NullString  `json:"mal"`
	NotifyMoe   db.NullString  `json:"notifymoe"`
	TMDB        db.NullString  `json:"tmdb"`
	TVDB        db.NullString  `json:"tvdb"`
	UpdatedAt   db.Timestamp   `json:"uat"`
}

type IdMapColumnStruct struct {
	Id          string
	Type        string
	AniDB       string
	AniList     string
	AniSearch   string
	AnimePlanet string
	IMDB        string
	Kitsu       string
	LiveChart   string
	MAL         string
	NotifyMoe   string
	TMDB        string
	TVDB        string
	UpdatedAt   string
}

var IdMapColumn = IdMapColumnStruct{
	Id:          "id",
	Type:        "type",
	AniDB:       "anidb",
	AniList:     "anilist",
	AniSearch:   "anisearch",
	AnimePlanet: "animeplanet",
	IMDB:        "imdb",
	Kitsu:       "kitsu",
	LiveChart:   "livechart",
	MAL:         "mal",
	NotifyMoe:   "notifymoe",
	TMDB:        "tmdb",
	TVDB:        "tvdb",
	UpdatedAt:   "uat",
}

var IdMapColumns = []string{
	IdMapColumn.Id,
	IdMapColumn.Type,
	IdMapColumn.AniDB,
	IdMapColumn.AniList,
	IdMapColumn.AniSearch,
	IdMapColumn.AnimePlanet,
	IdMapColumn.IMDB,
	IdMapColumn.Kitsu,
	IdMapColumn.LiveChart,
	IdMapColumn.MAL,
	IdMapColumn.NotifyMoe,
	IdMapColumn.TMDB,
	IdMapColumn.TVDB,
	IdMapColumn.UpdatedAt,
}

var query_get_id_map = fmt.Sprintf(
	"SELECT %s FROM %s WHERE %s IN ",
	strings.Join(IdMapColumns, ","),
	IdMapTableName,
	IdMapColumn.AniList,
)

func GetIdMapsForAniList(ids []int) ([]AnimeIdMap, error) {
	count := len(ids)
	query := query_get_id_map + "(" + util.RepeatJoin("?", count, ",") + ")"
	args := make([]any, count)
	for i := range ids {
		args[i] = strconv.Itoa(ids[i])
	}
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	idMaps := []AnimeIdMap{}
	for rows.Next() {
		var item rawAnimeIdMap
		if err := rows.Scan(
			&item.Id,
			&item.Type,
			&item.AniDB,
			&item.AniList,
			&item.AniSearch,
			&item.AnimePlanet,
			&item.IMDB,
			&item.Kitsu,
			&item.LiveChart,
			&item.MAL,
			&item.NotifyMoe,
			&item.TMDB,
			&item.TVDB,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		idMaps = append(idMaps, AnimeIdMap{
			Id:          item.Id,
			Type:        item.Type,
			AniList:     item.AniList.String,
			AniDB:       item.AniDB.String,
			AniSearch:   item.AniSearch.String,
			AnimePlanet: item.AnimePlanet.String,
			IMDB:        item.IMDB.String,
			Kitsu:       item.Kitsu.String,
			LiveChart:   item.LiveChart.String,
			MAL:         item.MAL.String,
			NotifyMoe:   item.NotifyMoe.String,
			TMDB:        item.TMDB.String,
			TVDB:        item.TVDB.String,
			UpdatedAt:   item.UpdatedAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return idMaps, nil
}

var query_get_type_by_anilist_ids = fmt.Sprintf(
	"SELECT %s, %s FROM %s WHERE %s IN ",
	IdMapColumn.AniList,
	IdMapColumn.Type,
	IdMapTableName,
	IdMapColumn.AniList,
)

func GetTypeByAnilistIds(ids []int) (map[int]AnimeIdMapType, error) {
	count := len(ids)
	if count == 0 {
		return nil, nil
	}

	query := query_get_type_by_anilist_ids + "(" + util.RepeatJoin("?", count, ",") + ")"
	args := make([]any, count)
	for i := range ids {
		args[i] = strconv.Itoa(ids[i])
	}
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	typeById := make(map[int]AnimeIdMapType, count)
	for rows.Next() {
		var id string
		var animeType AnimeIdMapType
		if err := rows.Scan(&id, &animeType); err != nil {
			return nil, err
		}
		if id, err := strconv.Atoi(id); err == nil {
			typeById[id] = animeType
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return typeById, nil
}

var query_get_type_by_kitsu_ids = fmt.Sprintf(
	"SELECT %s, %s FROM %s WHERE %s IN ",
	IdMapColumn.Kitsu,
	IdMapColumn.Type,
	IdMapTableName,
	IdMapColumn.Kitsu,
)

func GetTypeByKitsuIds(ids []int) (map[int]AnimeIdMapType, error) {
	count := len(ids)
	if count == 0 {
		return nil, nil
	}

	query := query_get_type_by_kitsu_ids + "(" + util.RepeatJoin("?", count, ",") + ")"
	args := make([]any, count)
	for i := range ids {
		args[i] = strconv.Itoa(ids[i])
	}
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	typeById := make(map[int]AnimeIdMapType, count)
	for rows.Next() {
		var id string
		var animeType AnimeIdMapType
		if err := rows.Scan(&id, &animeType); err != nil {
			return nil, err
		}
		if id, err := strconv.Atoi(id); err == nil {
			typeById[id] = animeType
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return typeById, nil
}

var query_get_anidb_id_by_kitsu_id = fmt.Sprintf(
	`SELECT im.%s, at.%s FROM %s im LEFT JOIN %s at ON at.%s = im.%s WHERE im.%s = ? LIMIT 1`,
	IdMapColumn.AniDB,
	anidb.TitleColumn.Season,
	IdMapTableName,
	anidb.TitleTableName,
	anidb.TitleColumn.TId,
	IdMapColumn.AniDB,
	IdMapColumn.Kitsu,
)

func GetAniDBIdByKitsuId(kitsuId string) (anidbId, season string, err error) {
	query := query_get_anidb_id_by_kitsu_id
	row := db.QueryRow(query, kitsuId)
	if err = row.Scan(&anidbId, &season); err != nil {
		if err == sql.ErrNoRows {
			return "", "", nil
		}
		return "", "", err
	}
	return anidbId, season, nil
}

var query_bulk_record_id_maps_before_values = fmt.Sprintf(
	`INSERT INTO %s AS aim (%s) VALUES `,
	IdMapTableName,
	strings.Join(IdMapColumns[1:len(IdMapColumns)-1], ","),
)
var query_bulk_record_id_maps_placeholder = "(" + util.RepeatJoin("?", len(IdMapColumns)-2, ",") + ")"
var query_bulk_record_id_maps_on_conflict_before_column = " ON CONFLICT ("
var query_bulk_record_id_maps_on_conflict_after_column = fmt.Sprintf(
	`) DO UPDATE SET %s = CASE WHEN aim.%s = '' THEN EXCLUDED.%s ELSE aim.%s END, %s = %s`,
	IdMapColumn.Type,
	IdMapColumn.Type,
	IdMapColumn.Type,
	IdMapColumn.Type,
	IdMapColumn.UpdatedAt,
	db.CurrentTimestamp,
)
var query_bulk_record_id_maps_on_conflict_set_by_column = map[string]string{
	IdMapColumn.AniDB:       fmt.Sprintf("%s = CASE WHEN aim.%s IS NULL THEN EXCLUDED.%s ELSE aim.%s END", IdMapColumn.AniDB, IdMapColumn.AniDB, IdMapColumn.AniDB, IdMapColumn.AniDB),
	IdMapColumn.AniList:     fmt.Sprintf("%s = CASE WHEN aim.%s IS NULL THEN EXCLUDED.%s ELSE aim.%s END", IdMapColumn.AniList, IdMapColumn.AniList, IdMapColumn.AniList, IdMapColumn.AniList),
	IdMapColumn.AniSearch:   fmt.Sprintf("%s = CASE WHEN aim.%s IS NULL THEN EXCLUDED.%s ELSE aim.%s END", IdMapColumn.AniSearch, IdMapColumn.AniSearch, IdMapColumn.AniSearch, IdMapColumn.AniSearch),
	IdMapColumn.AnimePlanet: fmt.Sprintf("%s = CASE WHEN aim.%s IS NULL THEN EXCLUDED.%s ELSE aim.%s END", IdMapColumn.AnimePlanet, IdMapColumn.AnimePlanet, IdMapColumn.AnimePlanet, IdMapColumn.AnimePlanet),
	IdMapColumn.IMDB:        fmt.Sprintf("%s = CASE WHEN aim.%s IS NULL THEN EXCLUDED.%s ELSE aim.%s END", IdMapColumn.IMDB, IdMapColumn.IMDB, IdMapColumn.IMDB, IdMapColumn.IMDB),
	IdMapColumn.Kitsu:       fmt.Sprintf("%s = CASE WHEN aim.%s IS NULL THEN EXCLUDED.%s ELSE aim.%s END", IdMapColumn.Kitsu, IdMapColumn.Kitsu, IdMapColumn.Kitsu, IdMapColumn.Kitsu),
	IdMapColumn.LiveChart:   fmt.Sprintf("%s = CASE WHEN aim.%s IS NULL THEN EXCLUDED.%s ELSE aim.%s END", IdMapColumn.LiveChart, IdMapColumn.LiveChart, IdMapColumn.LiveChart, IdMapColumn.LiveChart),
	IdMapColumn.MAL:         fmt.Sprintf("%s = CASE WHEN aim.%s IS NULL THEN EXCLUDED.%s ELSE aim.%s END", IdMapColumn.MAL, IdMapColumn.MAL, IdMapColumn.MAL, IdMapColumn.MAL),
	IdMapColumn.NotifyMoe:   fmt.Sprintf("%s = CASE WHEN aim.%s IS NULL THEN EXCLUDED.%s ELSE aim.%s END", IdMapColumn.NotifyMoe, IdMapColumn.NotifyMoe, IdMapColumn.NotifyMoe, IdMapColumn.NotifyMoe),
	IdMapColumn.TMDB:        fmt.Sprintf("%s = CASE WHEN aim.%s IS NULL THEN EXCLUDED.%s ELSE aim.%s END", IdMapColumn.TMDB, IdMapColumn.TMDB, IdMapColumn.TMDB, IdMapColumn.TMDB),
	IdMapColumn.TVDB:        fmt.Sprintf("%s = CASE WHEN aim.%s IS NULL THEN EXCLUDED.%s ELSE aim.%s END", IdMapColumn.TVDB, IdMapColumn.TVDB, IdMapColumn.TVDB, IdMapColumn.TVDB),
}

func normalizeOptionalId(id string) string {
	if id == "0" {
		return ""
	}
	return id
}

func getAnchorColumnValue(item AnimeIdMap, anchorColumnName string) string {
	switch anchorColumnName {
	case IdMapColumn.AniDB:
		return normalizeOptionalId(item.AniDB)
	case IdMapColumn.AniList:
		return normalizeOptionalId(item.AniList)
	case IdMapColumn.AniSearch:
		return normalizeOptionalId(item.AniSearch)
	case IdMapColumn.AnimePlanet:
		return normalizeOptionalId(item.AnimePlanet)
	case IdMapColumn.Kitsu:
		return normalizeOptionalId(item.Kitsu)
	case IdMapColumn.LiveChart:
		return normalizeOptionalId(item.LiveChart)
	case IdMapColumn.MAL:
		return normalizeOptionalId(item.MAL)
	case IdMapColumn.NotifyMoe:
		return normalizeOptionalId(item.NotifyMoe)
	default:
		panic("unsupported anchor column")
	}
}

func BulkRecordIdMaps(items []AnimeIdMap, anchorColumnName string) error {
	count := len(items)
	if count == 0 {
		return nil
	}

	var query strings.Builder
	query.WriteString(query_bulk_record_id_maps_before_values)

	seenMap := map[string]struct{}{}

	columnCount := len(IdMapColumns) - 2
	args := make([]any, 0, count*columnCount)
	for _, item := range items {
		anchorValue := getAnchorColumnValue(item, anchorColumnName)
		if _, seen := seenMap[anchorValue]; seen {
			count--
			continue
		}
		seenMap[anchorValue] = struct{}{}

		args = append(
			args,
			item.Type,
			db.NullString{String: normalizeOptionalId(item.AniDB)},
			db.NullString{String: normalizeOptionalId(item.AniList)},
			db.NullString{String: normalizeOptionalId(item.AniSearch)},
			db.NullString{String: normalizeOptionalId(item.AnimePlanet)},
			db.NullString{String: normalizeOptionalId(item.IMDB)},
			db.NullString{String: normalizeOptionalId(item.Kitsu)},
			db.NullString{String: normalizeOptionalId(item.LiveChart)},
			db.NullString{String: normalizeOptionalId(item.MAL)},
			db.NullString{String: normalizeOptionalId(item.NotifyMoe)},
			db.NullString{String: normalizeOptionalId(item.TMDB)},
			db.NullString{String: normalizeOptionalId(item.TVDB)},
		)
	}

	query.WriteString(util.RepeatJoin(query_bulk_record_id_maps_placeholder, count, ","))
	query.WriteString(query_bulk_record_id_maps_on_conflict_before_column)
	query.WriteString(anchorColumnName)
	query.WriteString(query_bulk_record_id_maps_on_conflict_after_column)
	for columnName, setColumnValue := range query_bulk_record_id_maps_on_conflict_set_by_column {
		if columnName == anchorColumnName {
			continue
		}
		query.WriteString(", ")
		query.WriteString(setColumnValue)
	}

	_, err := db.Exec(query.String(), args...)
	return err
}
