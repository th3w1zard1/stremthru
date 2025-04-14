package stremio_store

import (
	"net/url"
	"time"

	"github.com/MunifTanjim/stremthru/internal/cache"
	stremio_addon "github.com/MunifTanjim/stremthru/internal/stremio/addon"
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
