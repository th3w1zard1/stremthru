package stremio_wrap

import (
	"github.com/MunifTanjim/stremthru/stremio"
)

func getManifest(upstream *stremio.Manifest, ud *UserData) *stremio.Manifest {
	manifest := upstream
	manifest.ID = "st:wrap:" + manifest.ID
	manifest.Name = "StremThru(" + manifest.Name + ")"
	manifest.BehaviorHints = &stremio.BehaviorHints{
		Configurable:          true,
		ConfigurationRequired: !ud.HasRequiredValues(),
	}
	return manifest
}
