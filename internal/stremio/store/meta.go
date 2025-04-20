package stremio_store

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/MunifTanjim/stremthru/core"
	"github.com/MunifTanjim/stremthru/internal/cache"
	"github.com/MunifTanjim/stremthru/internal/shared"
	stremio_addon "github.com/MunifTanjim/stremthru/internal/stremio/addon"
	"github.com/MunifTanjim/stremthru/internal/torrent_info"
	"github.com/MunifTanjim/stremthru/internal/torrent_stream"
	"github.com/MunifTanjim/stremthru/internal/util"
	"github.com/MunifTanjim/stremthru/store"
	"github.com/MunifTanjim/stremthru/stremio"
)

var client = func() *stremio_addon.Client {
	return stremio_addon.NewClient(&stremio_addon.ClientConfig{})
}()

var cinemetaBaseUrl = func() *url.URL {
	url, err := url.Parse("https://v3-cinemeta.strem.io/")
	if err != nil {
		panic(err)
	}
	return url
}()

var metaCache = cache.NewCache[stremio.MetaHandlerResponse](&cache.CacheConfig{
	Lifetime: 10 * time.Minute,
	Name:     "stremio:store:catalog",
})

func fetchMeta(sType, imdbId, clientIp string) (stremio.MetaHandlerResponse, error) {
	var meta stremio.MetaHandlerResponse

	cacheKey := sType + ":" + imdbId
	if !metaCache.Get(cacheKey, &meta) {
		res, err := client.FetchMeta(&stremio_addon.FetchMetaParams{
			BaseURL:  cinemetaBaseUrl,
			Type:     sType,
			Id:       imdbId + ".json",
			ClientIP: clientIp,
		})
		if err != nil {
			return meta, err
		}
		meta = res.Data
		metaCache.Add(cacheKey, meta)
	}

	return meta, nil
}

func getPosterUrl(imdbId string) string {
	return "https://images.metahub.space/poster/small/" + imdbId + "/img"
}

func getMetaPreviewDescription(hash, name string) string {
	description := "[ 🧲 " + hash + " ]"

	r, err := util.ParseTorrentTitle(name)
	if err != nil {
		pttLog.Warn("failed to parse", "error", err, "title", name)
		return description
	}

	if r.Title != "" {
		description += " [ ✏️ " + r.Title + " ]"
	}
	if r.Year != "" || r.Date != "" {
		description += " [ 📅 "
		if r.Year != "" {
			description += r.Year
			if r.Date != "" {
				description += " | "
			}
		}
		if r.Date != "" {
			description += r.Date
		}
		description += " ]"
	}
	if r.Resolution != "" {
		description += " [ 🎥 " + r.Resolution + " ]"
	}
	if r.Quality != "" {
		description += " [ 💿 " + r.Quality + " ]"
	}
	if r.Codec != "" {
		description += " [ 🎞️ " + r.Codec + " ]"
	}
	if len(r.HDR) > 0 {
		description += " [ 📺 " + strings.Join(r.HDR, ",") + " ]"
	}
	if audioCount, channelCount := len(r.Audio), len(r.Channels); audioCount > 0 || channelCount > 0 {
		description += " [ 🎧 "
		if audioCount > 0 {
			description += strings.Join(r.Audio, ",")
			if channelCount > 0 {
				description += " | "
			}
		}
		if channelCount > 0 {
			description += strings.Join(r.Channels, ",")
		}
		description += " ]"
	}
	if r.ThreeD != "" {
		description += " [ 🎲 " + r.ThreeD + " ]"
	}
	if r.Network != "" {
		description += " [ 📡 " + r.Network + " ]"
	}
	if r.Group != "" {
		description += " [ ⚙️ " + r.Group + " ]"
	}
	if r.Site != "" {
		description += " [ 🔗 " + r.Site + " ]"
	}
	return description
}

func getStoreActionMeta(r *http.Request, storeCode string, eud string) stremio.Meta {
	released := time.Now().UTC()
	meta := stremio.Meta{
		Id:          getStoreActionId(storeCode),
		Type:        ContentTypeOther,
		Name:        "StremThru Store Actions",
		Description: "Actions for StremThru Store",
		Released:    released,
		Videos: []stremio.MetaVideo{
			{
				Id:       getStoreActionIdPrefix(storeCode) + "clear_cache",
				Title:    "Clear Cache",
				Released: released,
				Streams: []stremio.Stream{
					{
						URL:         ExtractRequestBaseURL(r).JoinPath("/stremio/store/" + eud + "/_/action/" + getStoreActionIdPrefix(storeCode) + "clear_cache").String(),
						Name:        "Clear Cache",
						Description: "Clear Cached Data for StremThru Store",
					},
				},
			},
		},
	}
	return meta
}

func handleMeta(w http.ResponseWriter, r *http.Request) {
	if !IsMethod(r, http.MethodGet) {
		shared.ErrorMethodNotAllowed(r).Send(w, r)
		return
	}

	if _, err := getContentType(r); err != nil {
		err.Send(w, r)
		return
	}

	ud, err := getUserData(r)
	if err != nil {
		SendError(w, r, err)
		return
	}

	id := getId(r)
	idr, err := parseId(id)
	if err != nil {
		SendError(w, r, err)
		return
	}

	idPrefix := getIdPrefix(idr.getStoreCode())

	if !strings.HasPrefix(id, idPrefix) {
		shared.ErrorBadRequest(r, "unsupported id: "+id).Send(w, r)
		return
	}

	ctx, err := ud.GetRequestContext(r, idr)
	if err != nil || ctx.Store == nil {
		if err != nil {
			LogError(r, "failed to get request context", err)
		}
		shared.ErrorBadRequest(r, "").Send(w, r)
		return
	}

	if id == getStoreActionId(idr.getStoreCode()) {
		eud, err := ud.GetEncoded()
		if err != nil {
			SendError(w, r, err)
			return
		}

		res := stremio.MetaHandlerResponse{
			Meta: getStoreActionMeta(r, idr.getStoreCode(), eud),
		}

		SendResponse(w, r, 200, res)
		return
	}

	params := &store.GetMagnetParams{
		Id:       strings.TrimPrefix(id, idPrefix),
		ClientIP: ctx.ClientIP,
	}
	params.APIKey = ctx.StoreAuthToken
	magnet, err := ctx.Store.GetMagnet(params)
	if err != nil {
		SendError(w, r, err)
		return
	}

	meta := stremio.Meta{
		Id:          id,
		Type:        ContentTypeOther,
		Name:        magnet.Name,
		Description: getMetaPreviewDescription(magnet.Hash, magnet.Name),
		Released:    magnet.AddedAt,
		Videos:      []stremio.MetaVideo{},
	}

	sType, sId := "", ""
	if stremIdByHashes, err := torrent_stream.GetStremIdByHashes([]string{magnet.Hash}); err != nil {
		log.Error("failed to get strem id by hashes", "error", err)
	} else {
		if sid := stremIdByHashes.Get(magnet.Hash); sid != "" {
			sid, _, isSeries := strings.Cut(sid, ":")
			sId = sid
			if isSeries {
				sType = "series"
			} else {
				sType = "movie"
			}
		}
	}

	metaVideoByKey := map[string]*stremio.MetaVideo{}
	if sId != "" {
		if r, err := fetchMeta(sType, sId, core.GetRequestIP(r)); err != nil {
			log.Error("failed to fetch meta", "error", err)
		} else {
			m := r.Meta
			meta.Description += " " + m.Description
			meta.Poster = m.Poster
			meta.Background = m.Background
			meta.Links = m.Links
			meta.Logo = m.Logo
			meta.Released = m.Released

			if sType == "series" {
				for i := range m.Videos {
					video := &m.Videos[i]
					key := strconv.Itoa(video.Season) + ":" + strconv.Itoa(video.Episode)
					metaVideoByKey[key] = video
				}
			}
		}
	}

	tInfo := torrent_info.TorrentInfoInsertData{
		Hash:         magnet.Hash,
		TorrentTitle: magnet.Name,
		Size:         magnet.Size,
		Source:       torrent_info.TorrentInfoSource(ctx.Store.GetName().Code()),
		Files:        []torrent_info.TorrentInfoInsertDataFile{},
	}

	tpttr, err := util.ParseTorrentTitle(magnet.Name)
	if err != nil {
		pttLog.Warn("failed to parse", "error", err, "title", magnet.Name)
	}

	for _, f := range magnet.Files {
		if !core.HasVideoExtension(f.Name) {
			continue
		}

		videoId := id + ":" + url.PathEscape(f.Link)
		video := stremio.MetaVideo{
			Id:        videoId,
			Title:     f.Name,
			Available: true,
			Released:  magnet.AddedAt,
		}

		season, episode := -1, -1
		pttr, err := util.ParseTorrentTitle(f.Name)
		if err != nil {
			pttLog.Warn("failed to parse", "error", err, "title", f.Name)
		} else {
			if len(pttr.Seasons) > 0 {
				season = pttr.Seasons[0]
				video.Season = season
			} else if len(tpttr.Seasons) == 1 {
				season = tpttr.Seasons[0]
				video.Season = season
			}
			if len(pttr.Episodes) > 0 {
				episode = pttr.Episodes[0]
				video.Episode = episode
			}
		}
		if season != -1 && episode != -1 {
			key := strconv.Itoa(season) + ":" + strconv.Itoa(episode)
			if sType == "series" {
				if metaVideo, ok := metaVideoByKey[key]; ok {
					video.Released = metaVideo.Released
					video.Thumbnail = metaVideo.Thumbnail
					video.Title = metaVideo.Name + "\n📄 " + f.Name
				} else {
					video.Title = pttr.Title + "\n📄 " + f.Name
				}
			}
		}

		meta.Videos = append(meta.Videos, video)
		tInfo.Files = append(tInfo.Files, torrent_info.TorrentInfoInsertDataFile{
			Name: f.Name,
			Idx:  f.Idx,
			Size: f.Size,
		})
	}

	go torrent_info.Upsert([]torrent_info.TorrentInfoInsertData{tInfo}, "", ctx.Store.GetName().Code() != store.StoreCodeRealDebrid)

	res := stremio.MetaHandlerResponse{
		Meta: meta,
	}

	SendResponse(w, r, 200, res)
}
