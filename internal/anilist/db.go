package anilist

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/MunifTanjim/stremthru/internal/anime"
	"github.com/MunifTanjim/stremthru/internal/db"
	"github.com/MunifTanjim/stremthru/internal/util"
)

const ListTableName = "anilist_list"

type AniListList struct {
	Id        string       `json:"id"`
	UpdatedAt db.Timestamp `json:"uat"`

	Medias []AniListMedia `json:"-"`
}

func (l *AniListList) GetURL() string {
	userName, name, ok := strings.Cut(l.Id, ":")
	if !ok {
		return ""
	}
	return "https://anilist.co/user/" + userName + "/animelist/" + name
}

func (l *AniListList) GetName() string {
	_, name, _ := strings.Cut(l.Id, ":")
	return name
}

func (l *AniListList) GetUserName() string {
	userName, _, _ := strings.Cut(l.Id, ":")
	return userName
}

func (l *AniListList) IsStale() bool {
	return time.Now().After(l.UpdatedAt.Add(12 * time.Hour))
}

type ListColumnStruct struct {
	Id        string
	UpdatedAt string
}

var ListColumn = ListColumnStruct{
	Id:        "id",
	UpdatedAt: "uat",
}

var ListColumns = []string{
	ListColumn.Id,
	ListColumn.UpdatedAt,
}

const MediaTableName = "anilist_media"

type genreList []string

func (genre genreList) Value() (driver.Value, error) {
	return json.Marshal(genre)
}

func (genre *genreList) Scan(value any) error {
	var bytes []byte
	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return errors.New("failed to convert value to []byte")
	}
	return json.Unmarshal(bytes, genre)
}

type AniListMedia struct {
	Id          int          `json:"id"`
	Type        string       `json:"type"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Banner      string       `json:"banner"`
	Cover       string       `json:"cover"`
	Duration    int          `json:"duration"`
	IsAdult     bool         `json:"is_adult"`
	StartYear   int          `json:"start_year"`
	UpdatedAt   db.Timestamp `json:"uat"`

	Genres genreList         `json:"-"`
	Score  int               `json:"-"`
	IdMap  *anime.AnimeIdMap `json:"-"`
}

func (m *AniListMedia) IsStale() bool {
	return time.Now().After(m.UpdatedAt.Add(5 * 24 * time.Hour))
}

type MediaColumnStruct struct {
	Id          string
	Type        string
	Title       string
	Description string
	Banner      string
	Cover       string
	Duration    string
	IsAdult     string
	StartYear   string
	UpdatedAt   string
}

var MediaColumn = MediaColumnStruct{
	Id:          "id",
	Type:        "type",
	Title:       "title",
	Description: "description",
	Banner:      "banner",
	Cover:       "cover",
	Duration:    "duration",
	IsAdult:     "is_adult",
	StartYear:   "start_year",
	UpdatedAt:   "uat",
}

var MediaColumns = []string{
	MediaColumn.Id,
	MediaColumn.Type,
	MediaColumn.Title,
	MediaColumn.Description,
	MediaColumn.Banner,
	MediaColumn.Cover,
	MediaColumn.Duration,
	MediaColumn.IsAdult,
	MediaColumn.StartYear,
	MediaColumn.UpdatedAt,
}

const ListMediaTableName = "anilist_list_media"

type AniListListMedia struct {
	ListId  string `json:"list_id"`
	MediaId int    `json:"media_id"`
	Score   int    `json:"score"`
}

type ListMediaColumnStruct struct {
	ListId  string
	MediaId string
	Score   string
}

var ListMediaColumn = ListMediaColumnStruct{
	MediaId: "media_id",
	ListId:  "list_id",
	Score:   "score",
}

var ListMediaColumns = []string{
	ListMediaColumn.ListId,
	ListMediaColumn.MediaId,
	ListMediaColumn.Score,
}

const MediaGenreTableName = "anilist_media_genre"

type AniListMediaGenre struct {
	MediaId int    `json:"media_id"`
	Genre   string `json:"genre"`
}

type MediaGenreColumnStruct struct {
	MediaId string
	Genre   string
}

var MediaGenreColumn = MediaGenreColumnStruct{
	MediaId: "media_id",
	Genre:   "genre",
}

var query_get_list_by_id = fmt.Sprintf(
	`SELECT %s FROM %s WHERE %s = ?`,
	db.JoinColumnNames(ListColumns...),
	ListTableName,
	ListColumn.Id,
)

func GetListById(id string) (*AniListList, error) {
	var list AniListList
	row := db.QueryRow(query_get_list_by_id, id)
	if err := row.Scan(&list.Id, &list.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	items, err := getListMedias(list.Id)
	if err != nil {
		return nil, err
	}
	list.Medias = items
	return &list, nil
}

var query_get_list_media_ids = fmt.Sprintf(
	`SELECT %s, %s FROM %s WHERE %s = ? ORDER BY %s DESC`,
	ListMediaColumn.MediaId,
	ListMediaColumn.Score,
	ListMediaTableName,
	ListMediaColumn.ListId,
	ListMediaColumn.Score,
)

func getListMediaIds(listId string) ([]int, map[int]int, error) {
	rows, err := db.Query(query_get_list_media_ids, listId)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	mediaIds := []int{}
	scoreByMediaId := map[int]int{}

	for rows.Next() {
		var mediaId int
		var score int
		if err := rows.Scan(&mediaId, &score); err != nil {
			return nil, nil, err
		}
		mediaIds = append(mediaIds, mediaId)
		scoreByMediaId[mediaId] = score
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return mediaIds, scoreByMediaId, nil
}

var query_get_medias = fmt.Sprintf(
	`SELECT %s, %s(mg.%s) AS genre FROM %s m LEFT JOIN %s mg ON m.%s = mg.%s WHERE m.%s IN `,
	db.JoinPrefixedColumnNames("m.", MediaColumns...),
	db.FnJSONGroupArray,
	MediaGenreColumn.Genre,
	MediaTableName,
	MediaGenreTableName,
	MediaColumn.Id,
	MediaGenreColumn.MediaId,
	MediaColumn.Id,
)
var query_get_medias_group_by = fmt.Sprintf(
	` GROUP BY m.%s`,
	MediaColumn.Id,
)

func getMedias(mediaIds []int, scoreByMediaId map[int]int) ([]AniListMedia, error) {
	count := len(mediaIds)
	if count == 0 {
		return nil, nil
	}

	query := query_get_medias + "(" + util.RepeatJoin("?", count, ",") + ")" + query_get_medias_group_by
	args := make([]any, count)
	for i := range mediaIds {
		args[i] = mediaIds[i]
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []AniListMedia
	for rows.Next() {
		var item AniListMedia
		if err := rows.Scan(
			&item.Id,
			&item.Type,
			&item.Title,
			&item.Description,
			&item.Banner,
			&item.Cover,
			&item.Duration,
			&item.IsAdult,
			&item.StartYear,
			&item.UpdatedAt,
			&item.Genres,
		); err != nil {
			return nil, err
		}
		if score, ok := scoreByMediaId[item.Id]; ok {
			item.Score = score
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	slices.SortFunc(items, func(a, b AniListMedia) int {
		return b.Score - a.Score
	})

	idMaps, err := anime.GetIdMapsForAniList(mediaIds)
	if err != nil {
		return nil, err
	}
	idMapById := map[string]*anime.AnimeIdMap{}
	for i := range idMaps {
		idMap := &idMaps[i]
		idMapById[idMap.AniList] = idMap
	}
	for i := range items {
		item := &items[i]
		if idMap, ok := idMapById[strconv.Itoa(item.Id)]; ok {
			item.IdMap = idMap
		}
	}

	return items, nil
}

func getListMedias(listId string) ([]AniListMedia, error) {
	mediaIds, scoreByMediaId, err := getListMediaIds(listId)
	if err != nil {
		return nil, err
	}
	return getMedias(mediaIds, scoreByMediaId)
}

var query_upsert_list = fmt.Sprintf(
	`INSERT INTO %s (%s) VALUES (?) ON CONFLICT (%s) DO UPDATE SET %s = %s`,
	ListTableName,
	ListColumn.Id,
	ListColumn.Id,
	ListColumn.UpdatedAt,
	db.CurrentTimestamp,
)

func UpsertList(list *AniListList) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
			return
		}
		tErr := tx.Rollback()
		err = errors.Join(tErr, err)
	}()

	_, err = tx.Exec(query_upsert_list, list.Id)
	if err != nil {
		return err
	}

	list.UpdatedAt = db.Timestamp{Time: time.Now()}

	err = upsertMedias(tx, list.Medias)
	if err != nil {
		return err
	}

	err = setListMedias(tx, list.Id, list.Medias)
	if err != nil {
		return err
	}

	return nil
}

var query_upsert_medias = fmt.Sprintf(
	`INSERT INTO %s (%s) VALUES `,
	MediaTableName,
	strings.Join(MediaColumns[0:len(MediaColumns)-1], ","),
)
var query_upsert_medias_values_placeholder = "(" + util.RepeatJoin("?", len(MediaColumns)-1, ",") + ")"
var query_upsert_medias_on_conflict = fmt.Sprintf(
	" ON CONFLICT (%s) DO UPDATE SET %s, %s = %s",
	MediaColumn.Id,
	strings.Join(
		[]string{
			fmt.Sprintf("%s = EXCLUDED.%s", MediaColumn.Type, MediaColumn.Type),
			fmt.Sprintf("%s = EXCLUDED.%s", MediaColumn.Title, MediaColumn.Title),
			fmt.Sprintf("%s = EXCLUDED.%s", MediaColumn.Description, MediaColumn.Description),
			fmt.Sprintf("%s = EXCLUDED.%s", MediaColumn.Banner, MediaColumn.Banner),
			fmt.Sprintf("%s = EXCLUDED.%s", MediaColumn.Cover, MediaColumn.Cover),
			fmt.Sprintf("%s = EXCLUDED.%s", MediaColumn.Duration, MediaColumn.Duration),
			fmt.Sprintf("%s = EXCLUDED.%s", MediaColumn.IsAdult, MediaColumn.IsAdult),
			fmt.Sprintf("%s = EXCLUDED.%s", MediaColumn.StartYear, MediaColumn.StartYear),
		},
		", ",
	),
	MediaColumn.UpdatedAt,
	db.CurrentTimestamp,
)

func upsertMedias(tx db.Executor, medias []AniListMedia) error {
	if len(medias) == 0 {
		return nil
	}

	for cMedias := range slices.Chunk(medias, 500) {
		count := len(cMedias)

		query := query_upsert_medias +
			util.RepeatJoin(query_upsert_medias_values_placeholder, count, ",") +
			query_upsert_medias_on_conflict

		columnCount := len(MediaColumns) - 1
		args := make([]any, count*columnCount)
		for i := range cMedias {
			media := &cMedias[i]
			args[i*columnCount+0] = media.Id
			args[i*columnCount+1] = media.Type
			args[i*columnCount+2] = media.Title
			args[i*columnCount+3] = media.Description
			args[i*columnCount+4] = media.Banner
			args[i*columnCount+5] = media.Cover
			args[i*columnCount+6] = media.Duration
			args[i*columnCount+7] = media.IsAdult
			args[i*columnCount+8] = media.StartYear
		}

		_, err := tx.Exec(query, args...)
		if err != nil {
			return err
		}

		for _, media := range cMedias {
			err = setMediaGenre(tx, media.Id, media.Genres)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

var query_set_media_genre_before_values = fmt.Sprintf(
	`INSERT INTO %s (%s, %s) VALUES `,
	MediaGenreTableName,
	MediaGenreColumn.MediaId,
	MediaGenreColumn.Genre,
)
var query_set_media_genre_values_placeholder = "(?, ?)"
var query_set_media_genre_after_values = fmt.Sprintf(
	` ON CONFLICT DO NOTHING`,
)
var query_cleanup_media_genre = fmt.Sprintf(
	`DELETE FROM %s WHERE %s = ? AND %s NOT IN `,
	MediaGenreTableName,
	MediaGenreColumn.MediaId,
	MediaGenreColumn.Genre,
)

func setMediaGenre(tx db.Executor, mediaId int, genres []string) error {
	count := len(genres)

	cleanupArgs := make([]any, 1+count)
	cleanupArgs[0] = mediaId
	for i, genre := range genres {
		cleanupArgs[1+i] = genre
	}
	cleanupQuery := query_cleanup_media_genre + "(" + util.RepeatJoin("?", count, ",") + ")"
	if _, err := tx.Exec(cleanupQuery, cleanupArgs...); err != nil {
		return err
	}

	if count == 0 {
		return nil
	}

	query := query_set_media_genre_before_values +
		util.RepeatJoin(query_set_media_genre_values_placeholder, count, ",") +
		query_set_media_genre_after_values
	args := make([]any, count*2)
	for i, genre := range genres {
		args[i*2] = mediaId
		args[i*2+1] = genre
	}

	if _, err := tx.Exec(query, args...); err != nil {
		return err
	}

	return nil
}

var query_set_list_medias_before_values = fmt.Sprintf(
	`INSERT INTO %s (%s, %s, %s) VALUES `,
	ListMediaTableName,
	ListMediaColumn.ListId,
	ListMediaColumn.MediaId,
	ListMediaColumn.Score,
)
var query_set_list_medias_values_placeholder = "(?,?,?)"
var query_set_list_medias_after_values = fmt.Sprintf(
	` ON CONFLICT (%s, %s) DO UPDATE SET %s = EXCLUDED.%s`,
	ListMediaColumn.ListId,
	ListMediaColumn.MediaId,
	ListMediaColumn.Score,
	ListMediaColumn.Score,
)
var query_cleanup_list_medias = fmt.Sprintf(
	`DELETE FROM %s WHERE %s = ? AND %s NOT IN `,
	ListMediaTableName,
	ListMediaColumn.ListId,
	ListMediaColumn.MediaId,
)

func setListMedias(tx *db.Tx, listId string, medias []AniListMedia) error {
	count := len(medias)

	cleanupArgs := make([]any, 1+count)
	cleanupArgs[0] = listId
	for i := range medias {
		cleanupArgs[1+i] = medias[i].Id
	}
	cleanupQuery := query_cleanup_list_medias + "(" + util.RepeatJoin("?", count, ",") + ")"
	if _, err := tx.Exec(cleanupQuery, cleanupArgs...); err != nil {
		return err
	}

	if count == 0 {
		return nil
	}

	query := query_set_list_medias_before_values +
		util.RepeatJoin(query_set_list_medias_values_placeholder, count, ",") +
		query_set_list_medias_after_values
	args := make([]any, count*3)
	for i := range medias {
		item := &medias[i]
		args[i*3+0] = listId
		args[i*3+1] = item.Id
		args[i*3+2] = item.Score
	}

	if _, err := tx.Exec(query, args...); err != nil {
		return err
	}

	return nil
}
