package stremio_store

import (
	"errors"
	"strings"

	"github.com/MunifTanjim/stremthru/store"
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

type ParsedId struct {
	storeCode    store.StoreCode
	storeName    store.StoreName
	isDeprecated bool
	isST         bool
}

func (idr ParsedId) getStoreCode() string {
	if idr.isST {
		if idr.storeCode == "" {
			return "st"
		}
		if idr.isDeprecated {
			return "st:" + string(idr.storeCode)
		}
		return "st-" + string(idr.storeCode)
	}
	return string(idr.storeCode)
}

func parseId(id string) (*ParsedId, error) {
	parts := strings.SplitN(id, ":", 5)
	count := len(parts)
	if count < 3 {
		return nil, errors.New("invalid id")
	}

	r := ParsedId{}
	storeCode := parts[2]
	if strings.Contains(storeCode, "-") {
		scParts := strings.Split(storeCode, "-")
		if scParts[0] == "st" {
			r.isST = true
			r.storeCode = store.StoreCode(scParts[1])
		} else {
			r.storeCode = store.StoreCode(scParts[0])
		}
	} else {
		switch storeCode {
		case "st":
			r.isST = true
			r.isDeprecated = true
			if count > 3 {
				r.storeCode = store.StoreCode(parts[3])
				if r.storeCode.Name() == "" {
					r.storeCode = ""
				}
			}
		default:
			r.storeCode = store.StoreCode(parts[2])
		}
	}

	r.storeName = r.storeCode.Name()

	return &r, nil
}
