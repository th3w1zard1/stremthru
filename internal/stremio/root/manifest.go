package stremio_root

import (
	"net/http"

	"github.com/MunifTanjim/stremthru/internal/config"
	"github.com/MunifTanjim/stremthru/internal/shared"
	"github.com/MunifTanjim/stremthru/stremio"
)

func getManifest(r *http.Request) *stremio.Manifest {
	manifest := &stremio.Manifest{
		AddonCatalogs: []stremio.Catalog{
			{Type: "all", Id: "stremthru", Name: "StremThru"},
		},
		BehaviorHints: &stremio.BehaviorHints{
			Configurable: true,
		},
		Catalogs:    []stremio.Catalog{},
		Description: "Companion for Stremio",
		ID:          shared.GetReversedHostname(r),
		Logo:        "https://emojiapi.dev/api/v1/sparkles/256.png",
		Name:        "StremThru",
		Resources:   []stremio.Resource{{Name: stremio.ResourceNameAddonCatalog}},
		Types:       []stremio.ContentType{},
		Version:     config.Version,
	}

	if manifest.ID == "com.elfhosted.stremthru" {
		manifest.StremioAddonsConfig = &stremio.StremioAddonsConfig{
			Issuer:    "https://stremio-addons.net",
			Signature: "eyJhbGciOiJkaXIiLCJlbmMiOiJBMTI4Q0JDLUhTMjU2In0..FZwjVa6u7prcA4BF4WPs0A.AdUxPjZHiLXZRe4VMtGOd1wUBcG9effo9zQjXsc2eJ2mu6QLJ1kYC70uPUWqZZjTYdcC23kRnI1hn2JwTFddVSwXsUHENeRstFI3FpxRXx2B3_bpqDKiKJeICo8zMbm6.X9hnbnVUkDaYjrCQBLmzrA",
		}
	}

	return manifest
}
