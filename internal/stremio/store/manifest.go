package stremio_store

import (
	"strings"

	"github.com/MunifTanjim/stremthru/core"
	"github.com/MunifTanjim/stremthru/internal/config"
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

func getManifest(ud *UserData) *stremio.Manifest {
	name := "Store"
	description := "StremThru Store Catalog and Search"
	storeName := ""
	storeCode := ""
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

	id := "dev.muniftanjim.stremthru.store." + storeCode

	idPrefix := getIdPrefix(storeCode)

	contactEmail, _ := core.Base64Decode("Z2l0aHViQG11bmlmdGFuamltLmRldg==")
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
		ContactEmail: contactEmail,
		BehaviorHints: &stremio.BehaviorHints{
			Configurable:          true,
			ConfigurationRequired: !ud.HasRequiredValues(),
		},
	}

	return manifest
}
