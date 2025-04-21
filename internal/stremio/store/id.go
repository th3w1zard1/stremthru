package stremio_store

import (
	"errors"
	"strings"

	"github.com/MunifTanjim/stremthru/store"
)

func isStoreId(id string) bool {
	return strings.HasPrefix(id, "st:store:")
}

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
	isUsenet     bool
	isDeprecated bool
	isST         bool
	code         string
}

func (idr *ParsedId) getStoreCode() string {
	if idr.code == "" {
		if idr.isST {
			if idr.storeCode == "" {
				idr.code = "st"
			} else if idr.isDeprecated {
				idr.code = "st:" + string(idr.storeCode)
			} else {
				idr.code = "st-" + string(idr.storeCode)
				if idr.isUsenet {
					idr.code += "-usenet"
				}
			}
		} else {
			idr.code = string(idr.storeCode)
			if idr.isUsenet {
				idr.code += "-usenet"
			}
		}
	}
	return idr.code
}

func (idr *ParsedId) getIdPrefix() string {
	return getIdPrefix(idr.getStoreCode())
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
			if len(scParts) > 2 && scParts[2] == "usenet" {
				r.isUsenet = true
			}
		} else {
			r.storeCode = store.StoreCode(scParts[0])
			if len(scParts) > 1 && scParts[1] == "usenet" {
				r.isUsenet = true
			}
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
