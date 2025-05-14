package stremio_shared

import (
	"github.com/MunifTanjim/stremthru/internal/config"
	"github.com/MunifTanjim/stremthru/internal/stremio/configure"
)

func GetStoreCodeOptions() []configure.ConfigOption {
	options := []configure.ConfigOption{
		{Value: "", Label: "StremThru"},
		{Value: "ad", Label: "AllDebrid"},
		{Value: "dl", Label: "DebridLink"},
		{Value: "ed", Label: "EasyDebrid"},
		{Value: "oc", Label: "Offcloud"},
		{Value: "pm", Label: "Premiumize"},
		{Value: "pp", Label: "PikPak"},
		{Value: "rd", Label: "RealDebrid"},
		{Value: "tb", Label: "TorBox"},
	}
	if config.IsPublicInstance {
		options[0].Disabled = true
		options[0].Label = ""
	}
	return options
}
