package buddy

import (
	"log"
	"time"

	"github.com/MunifTanjim/stremthru/core"
	"github.com/MunifTanjim/stremthru/internal/cache"
	"github.com/MunifTanjim/stremthru/internal/config"
	"github.com/MunifTanjim/stremthru/store"
)

var Client = NewAPIClient(&APIClientConfig{
	BaseURL: config.BuddyURL,
	APIKey:  config.BuddyAuthToken,
})

// If Buddy is available, using this cache to save Buddy's response.
// So that we don't call Buddy too frequently.
//
// If Buddy is unavailable, using this cache as a local store.
var checkMagnetCache = func() cache.Cache[store.CheckMagnetDataItem] {
	lifetime := 10 * time.Minute
	if !Client.IsAvailable() {
		lifetime = 24 * time.Hour
	}
	return cache.NewCache[store.CheckMagnetDataItem](&cache.CacheConfig{
		Name:     "buddy:checkMagnet",
		Lifetime: lifetime,
	})
}()

func getCachedCheckMagnet(s store.Store, magnetHash string) *store.CheckMagnetDataItem {
	v := &store.CheckMagnetDataItem{}
	if ok := checkMagnetCache.Get(string(s.GetName())+":"+magnetHash, v); ok {
		return v
	}
	return nil
}

func setCachedCheckMagnet(s store.Store, magnetHash string, v *store.CheckMagnetDataItem) {
	if v == nil {
		checkMagnetCache.Remove(string(s.GetName()) + ":" + magnetHash)
	} else {
		checkMagnetCache.Add(string(s.GetName())+":"+magnetHash, *v)
	}
}

func trackMagnetWithoutBuddy(s store.Store, hash string, files []store.MagnetFile, cacheMiss bool) {
	magnet, err := core.ParseMagnetLink(hash)
	if err != nil {
		log.Printf("failed to track magnet cache for %s:%s: %v\n", s.GetName(), hash, err)
		return
	}
	status := store.MagnetStatusUnknown
	if cacheMiss {
		status = store.MagnetStatusCached
	}
	setCachedCheckMagnet(s, hash, &store.CheckMagnetDataItem{
		Hash:   magnet.Hash,
		Magnet: magnet.Link,
		Status: status,
		Files:  files,
	})
}

func TrackMagnet(s store.Store, hash string, files []store.MagnetFile, cacheMiss bool) {
	if !Client.IsAvailable() {
		trackMagnetWithoutBuddy(s, hash, files, cacheMiss)
		return
	}

	if _, err := Client.TrackMagnetCache(&TrackMagnetCacheParams{
		Store:     s.GetName(),
		Hash:      hash,
		Files:     files,
		CacheMiss: cacheMiss,
	}); err != nil {
		log.Printf("failed to track magnet cache for %s:%s: %v\n", s.GetName(), hash, err)
	}
}

func checkMagnetWithoutBuddy(s store.Store, hashes []string) (*store.CheckMagnetData, error) {
	data := &store.CheckMagnetData{
		Items: []store.CheckMagnetDataItem{},
	}

	uncachedHashes := []string{}
	for _, hash := range hashes {
		if v := getCachedCheckMagnet(s, hash); v != nil && v.Status == store.MagnetStatusCached {
			data.Items = append(data.Items, *v)
		} else {
			uncachedHashes = append(uncachedHashes, hash)
		}
	}

	if len(uncachedHashes) == 0 {
		return data, nil
	}

	for _, hash := range uncachedHashes {
		magnet, err := core.ParseMagnetLink(hash)
		if err != nil {
			return nil, err
		}
		item := store.CheckMagnetDataItem{
			Hash:   magnet.Hash,
			Magnet: magnet.Link,
			Status: store.MagnetStatusUnknown,
			Files:  []store.MagnetFile{},
		}
		data.Items = append(data.Items, item)
	}

	return data, nil
}

func CheckMagnet(s store.Store, hashes []string) (*store.CheckMagnetData, error) {
	if !Client.IsAvailable() {
		return checkMagnetWithoutBuddy(s, hashes)
	}

	data := &store.CheckMagnetData{
		Items: []store.CheckMagnetDataItem{},
	}

	uncachedHashes := []string{}
	for _, hash := range hashes {
		if v := getCachedCheckMagnet(s, hash); v != nil {
			data.Items = append(data.Items, *v)
		} else {
			uncachedHashes = append(uncachedHashes, hash)
		}
	}

	if len(uncachedHashes) == 0 {
		return data, nil
	}

	res, err := Client.CheckMagnetCache(&CheckMagnetCacheParams{
		Store:  s.GetName(),
		Hashes: uncachedHashes,
	})
	if err != nil {
		return nil, err
	}
	for _, item := range res.Data.Items {
		setCachedCheckMagnet(s, item.Hash, &item)
		data.Items = append(data.Items, item)
	}

	return data, nil
}
