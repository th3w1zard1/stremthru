package config

import "strings"

type stremioConfigTorz struct {
	LazyPull bool
}

type StremioConfig struct {
	Torz stremioConfigTorz
}

func parseStremio() StremioConfig {
	torzLazyPull := strings.ToLower(getEnv("STREMTHRU_STREMIO_TORZ_LAZY_PULL"))
	stremio := StremioConfig{
		Torz: stremioConfigTorz{
			LazyPull: torzLazyPull == "true",
		},
	}
	return stremio
}

var Stremio = parseStremio()
