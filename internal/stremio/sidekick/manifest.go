package stremio_sidekick

import (
	"net/http"

	"github.com/MunifTanjim/stremthru/internal/config"
	"github.com/MunifTanjim/stremthru/internal/shared"
	"github.com/MunifTanjim/stremthru/stremio"
)

func GetManifest(r *http.Request) *stremio.Manifest {
	return &stremio.Manifest{
		ID:          shared.GetReversedHostname(r) + ".sidekick",
		Name:        "Stremio Sidekick",
		Description: "Extra Features for Stremio",
		Version:     config.Version,
		Resources:   []stremio.Resource{},
		Types:       []stremio.ContentType{},
		Catalogs:    []stremio.Catalog{},
		BehaviorHints: &stremio.BehaviorHints{
			Configurable:          true,
			ConfigurationRequired: true,
		},
	}
}
