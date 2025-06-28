package stremio_shared

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/MunifTanjim/go-ptt"
	"github.com/MunifTanjim/stremthru/internal/anime"
	"github.com/MunifTanjim/stremthru/internal/torrent_info"
	"github.com/MunifTanjim/stremthru/internal/torrent_stream"
	"github.com/MunifTanjim/stremthru/internal/util"
	"github.com/MunifTanjim/stremthru/store"
)

func MatchFileByIdx(files []store.MagnetFile, idx int, storeCode store.StoreCode) *store.MagnetFile {
	if idx == -1 || storeCode != store.StoreCodeRealDebrid {
		return nil
	}
	for i := range files {
		f := &files[i]
		if f.Idx == idx {
			return f
		}
	}
	return nil
}

func MatchFileByLargestSize(files []store.MagnetFile) (file *store.MagnetFile) {
	for i := range files {
		f := &files[i]
		if file == nil || file.Size < f.Size {
			file = f
		}
	}
	return file
}

func MatchFileByName(files []store.MagnetFile, name string) *store.MagnetFile {
	if name == "" {
		return nil
	}
	for i := range files {
		f := &files[i]
		if f.Name == name {
			return f
		}
	}
	return nil
}

func MatchFileByPattern(files []store.MagnetFile, pattern *regexp.Regexp) *store.MagnetFile {
	if pattern == nil {
		return nil
	}
	for i := range files {
		f := &files[i]
		if pattern.MatchString(f.Name) {
			return f
		}
	}
	return nil
}

var parse_season_episode = ptt.GetPartialParser([]string{"releaseType", "seasons", "episodes"})

var digits_regex = regexp.MustCompile(`\b(\d+)\b`)

type seasonEpisodeData struct {
	season  int
	episode int
}

func getSeasonEpisode(title string, extractDigitsAsEpisodeAgressively bool) seasonEpisodeData {
	data := seasonEpisodeData{-1, -1}
	r := parse_season_episode(title)
	if err := r.Error(); err != nil {
		log.Error("failed to parse season episode", "title", title, "error", err)
		return data
	}
	if len(r.Seasons) > 0 {
		data.season = r.Seasons[0]
	}
	if len(r.Episodes) > 0 {
		data.episode = r.Episodes[0]
	}
	if extractDigitsAsEpisodeAgressively && data.season == -1 && data.episode == -1 {
		matches := digits_regex.FindAllString(title, 2)
		if len(matches) == 1 {
			data.episode, _ = strconv.Atoi(matches[0])
		}
	}
	return data
}

func MatchFileByStremId(files []store.MagnetFile, sid string, magnetHash string, storeCode store.StoreCode) *store.MagnetFile {
	if f, err := torrent_stream.GetFile(magnetHash, sid); err != nil {
		log.Error("failed to get file by strem id", "hash", magnetHash, "sid", sid, "error", err)
	} else if f != nil {
		if file := MatchFileByIdx(files, f.Idx, storeCode); file != nil {
			log.Debug("matched file by strem id - fileidx", "hash", magnetHash, "sid", sid, "filename", file.Name, "fileidx", file.Idx, "store", storeCode)
			return file
		}
		if file := MatchFileByName(files, f.Name); file != nil {
			log.Debug("matched file by strem id - filename", "hash", magnetHash, "sid", sid, "filename", file.Name, "fileidx", file.Idx, "store", storeCode)
			return file
		}
	}
	if strings.HasPrefix(sid, "kitsu:") {
		kitsuId, episode, _ := strings.Cut(strings.TrimPrefix(sid, "kitsu:"), ":")
		_, season, err := anime.GetAniDBIdByKitsuId(kitsuId)
		if err != nil {
			log.Error("failed to get anidb id by kitsu id", "error", err, "kitsu_id", kitsuId)
			return nil
		}
		tInfo, err := torrent_info.GetByHash(magnetHash)
		if err != nil {
			log.Error("failed to get torrent info by hash", "error", err, "hash", magnetHash)
			return nil
		}
		expectedEpisode := util.SafeParseInt(episode, -1)
		expectedSeason := util.SafeParseInt(season, -1)

		filesForSeason := []*store.MagnetFile{}
		dataByIdx := map[int]seasonEpisodeData{}

		minEpisode := 99999
		for i := range files {
			f := &files[i]
			d := getSeasonEpisode(f.Name, true)
			if (d.episode != -1) && ((d.season == -1 && expectedSeason == 1) || d.season == expectedSeason) {
				filesForSeason = append(filesForSeason, f)
				idx := len(filesForSeason) - 1
				dataByIdx[idx] = d
				if d.episode < minEpisode {
					minEpisode = d.episode
				}
			}
		}

		for i, f := range filesForSeason {
			d := dataByIdx[i]
			if d.episode == expectedEpisode || (len(tInfo.Episodes) == 0 && minEpisode > 1 && d.episode-minEpisode+1 == expectedEpisode) {
				return f
			}
		}
		return nil
	}
	if parts := strings.SplitN(sid, ":", 3); len(parts) == 3 {
		expectedSeason, err := strconv.Atoi(parts[1])
		if err != nil {
			log.Warn("failed to parse season from strem id", "sid", sid, "error", err)
			return nil
		}
		expectedEpisode, err := strconv.Atoi(parts[2])
		if err != nil {
			log.Warn("failed to parse episode from strem id", "sid", sid, "error", err)
			return nil
		}
		for i := range files {
			f := &files[i]
			if d := getSeasonEpisode(f.Name, false); d.season == expectedSeason && d.episode == expectedEpisode {
				return f
			}
		}
	}
	return nil
}
