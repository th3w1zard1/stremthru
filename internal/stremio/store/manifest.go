package stremio_store

import (
	"net/http"
	"strings"

	"github.com/MunifTanjim/stremthru/internal/config"
	"github.com/MunifTanjim/stremthru/internal/shared"
	"github.com/MunifTanjim/stremthru/store"
	"github.com/MunifTanjim/stremthru/stremio"
)

func getCatalogId(storeCode string) string {
	return "st:store:" + storeCode
}

func getIdPrefix(storeCode string) string {
	return getCatalogId(storeCode) + ":"
}

func getStoreActionId(storeCode string) string {
	return getIdPrefix(storeCode) + "action"
}

func getStoreActionIdPrefix(storeCode string) string {
	return getStoreActionId(storeCode) + ":"
}

const ContentTypeOther = "other"

const (
	CatalogGenreVideo     = "Video"
	CatalogGenreStremThru = "StremThru"
)

func GetManifest(r *http.Request, ud *UserData) *stremio.Manifest {
	isConfigured := ud.HasRequiredValues()

	id := shared.GetReversedHostname(r) + ".store"
	name := "Store"
	description := "Explore and Search Store Catalog"
	storeName := ""
	storeCode := ""
	if isConfigured {
		switch ud.StoreName {
		case "":
			storeName = "StremThru"
			storeCode = "st"
		case "stremthru":
			storeName = "StremThru"
			storeCode = "st"
		default:
			storeName = string(store.StoreName(ud.StoreName))
			storeCode = string(store.StoreName(ud.StoreName).Code())
		}

		name = name + " | " + strings.ToUpper(storeCode)
		description = description + " - " + storeName
	} else {
		name = "StremThru Store"
	}

	if storeCode != "" {
		id += "." + storeCode
	}

	idPrefix := getIdPrefix(storeCode)

	manifest := &stremio.Manifest{
		ID:          id,
		Name:        name,
		Description: description,
		Version:     config.Version,
		Resources: []stremio.Resource{
			{
				Name:       stremio.ResourceNameMeta,
				Types:      []stremio.ContentType{ContentTypeOther},
				IDPrefixes: []string{idPrefix},
			},
			{
				Name:       stremio.ResourceNameStream,
				Types:      []stremio.ContentType{ContentTypeOther},
				IDPrefixes: []string{idPrefix},
			},
		},
		Types: []stremio.ContentType{},
		Catalogs: []stremio.Catalog{
			{
				Id:   getCatalogId(storeCode),
				Name: "Store",
				Type: ContentTypeOther,
				Extra: []stremio.CatalogExtra{
					{
						Name: "search",
					},
					{
						Name: "skip",
					},
					{
						Name:    "genre",
						Options: []string{CatalogGenreVideo, CatalogGenreStremThru},
					},
				},
			},
		},
		BehaviorHints: &stremio.BehaviorHints{
			Configurable:          true,
			ConfigurationRequired: !isConfigured,
		},
	}

	return manifest
}
