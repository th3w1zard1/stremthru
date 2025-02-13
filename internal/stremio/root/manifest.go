package stremio_root

import (
	"net/http"

	"github.com/MunifTanjim/stremthru/internal/config"
	"github.com/MunifTanjim/stremthru/stremio"
)

func getManifest(r *http.Request) *stremio.Manifest {
	manifest := &stremio.Manifest{
		AddonCatalogs: []stremio.Catalog{{Type: "other", Id: "stremthru", Name: "StremThru"}},
		Background:    "",
		BehaviorHints: &stremio.BehaviorHints{},
		Catalogs:      []stremio.Catalog{},
		Description:   "Companion for Stremio",
		ID:            "stremthru",
		Logo:          "",
		Name:          "StremThru",
		Resources:     []stremio.Resource{{Name: stremio.ResourceNameAddonCatalog}},
		Types:         []stremio.ContentType{},
		Version:       config.Version,
	}

	return manifest
}
