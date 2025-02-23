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

	return manifest
}
