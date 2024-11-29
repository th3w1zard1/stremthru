package buddy

import (
	"log"

	"github.com/MunifTanjim/stremthru/internal/config"
	"github.com/MunifTanjim/stremthru/store"
)

var Client = NewAPIClient(&APIClientConfig{
	BaseURL: config.BuddyURL,
})

func TrackMagnetCache(s store.Store, hash string, files []store.MagnetFile, cacheMiss bool) {
	if !Client.IsAvailable() {
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

func CheckMagnetCache(s store.Store, hashes []string) (*store.CheckMagnetData, error) {
	if Client.IsAvailable() {
		res, err := Client.CheckMagnetCache(&CheckMagnetCacheParams{
			Store:  s.GetName(),
			Hashes: hashes,
		})
		if err != nil {
			return nil, err
		}
		return &res.Data, nil
	}
	return nil, nil
}
