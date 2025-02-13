package magnet_cache

import (
	"strings"

	"github.com/MunifTanjim/stremthru/internal/db"
	"github.com/MunifTanjim/stremthru/internal/logger"
	"github.com/MunifTanjim/stremthru/internal/util"
)

const FileTableName = "magnet_cache_file"

var mcfLog = logger.Scoped(FileTableName)

type MagnetCacheFile struct {
	Hash string `json:"-"`
	Name string `json:"n"`
	Idx  int    `json:"i"`
	Size int    `json:"s"`
	SId  string `json:"-"`
}

func GetFilesByHashes(hashes []string) (map[string]Files, error) {
	byHash := map[string]Files{}

	if len(hashes) == 0 {
		return byHash, nil
	}

	args := make([]interface{}, len(hashes))
	hashPlaceholders := make([]string, len(hashes))
	for i, hash := range hashes {
		args[i] = hash
		hashPlaceholders[i] = "?"
	}

	rows, err := db.Query("SELECT h, "+db.FnJSONGroupArray+"("+db.FnJSONObject+"('i', i, 'n', n, 's', s)) AS files FROM "+FileTableName+" WHERE h IN ("+strings.Join(hashPlaceholders, ",")+") GROUP BY h", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		hash := ""
		files := Files{}
		if err := rows.Scan(&hash, &files); err != nil {
			return nil, err
		}
		byHash[hash] = files
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return byHash, nil
}

func TrackFiles(hash string, files Files, sid string, discardIdx bool) {
	if len(files) == 0 {
		return
	}

	var query strings.Builder
	query.WriteString("INSERT INTO " + FileTableName + " (h,i,n,s,sid) VALUES ")
	placeholder := "(?,?,?,?,?)"
	count := 0

	var args []interface{}

	for _, file := range files {
		if count > 0 {
			query.WriteString(",")
		}
		query.WriteString(placeholder)
		fsid := file.SId
		if fsid == "" {
			fsid = sid
		}
		idx := file.Idx
		if discardIdx {
			idx = -1
		}
		args = append(args, hash, idx, file.Name, file.Size, fsid)
		count++
	}

	query.WriteString(" ON CONFLICT (h, n) DO UPDATE SET i = CASE WHEN " + FileTableName + ".i = -1 THEN excluded.i ELSE " + FileTableName + ".i END, s = CASE WHEN " + FileTableName + ".s = -1 THEN excluded.s ELSE " + FileTableName + ".s END, sid = CASE WHEN " + FileTableName + ".sid IN ('', '*') THEN excluded.sid ELSE " + FileTableName + ".sid END")
	_, err := db.Exec(query.String(), args...)
	if err != nil {
		mcfLog.Error("failed to track", "error", err)
	}
}

func execBulkTrackFiles(count int, args []any) {
	var query strings.Builder
	query.WriteString("INSERT INTO " + FileTableName + " (h,i,n,s,sid) VALUES ")

	placeholder := "(?,?,?,?,?)"
	if count > 0 {
		query.WriteString(util.RepeatJoin(placeholder, count, ","))
		query.WriteString(" ON CONFLICT (h, n) DO UPDATE SET i = CASE WHEN " + FileTableName + ".i = -1 THEN EXCLUDED.i ELSE " + FileTableName + ".i END, s = CASE WHEN " + FileTableName + ".s = -1 THEN EXCLUDED.s ELSE " + FileTableName + ".s END, sid = CASE WHEN " + FileTableName + ".sid IN ('', '*') THEN EXCLUDED.sid ELSE " + FileTableName + ".sid END")
		_, err := db.Exec(query.String(), args...)
		if err != nil {
			mcfLog.Error("failed to partially bulk track", "error", err)
		}
	}
}

func BulkTrackFiles(filesByHash map[string]Files, sid string, discardIdx bool) {
	count := 0
	args := []any{}
	for hash, files := range filesByHash {
		for _, file := range files {
			fsid := file.SId
			if fsid == "" {
				fsid = sid
			}
			idx := file.Idx
			if discardIdx {
				idx = -1
			}
			args = append(args, hash, idx, file.Name, file.Size, fsid)
			count++
		}
		if count >= 200 {
			execBulkTrackFiles(count, args)
			count = 0
			args = []any{}
		}
	}
	execBulkTrackFiles(count, args)
}
