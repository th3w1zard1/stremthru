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
	storeCode store.StoreCode
	storeName store.StoreName
	isST      bool
}

func (idr ParsedId) getStoreCode() string {
	if idr.isST {
		if idr.storeCode == "" {
			return "st"
		}
		return "st:" + string(idr.storeCode)
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
	switch parts[2] {
	case "st":
		r.isST = true
		if count > 3 {
			r.storeCode = store.StoreCode(parts[3])
		}
	default:
		r.storeCode = store.StoreCode(parts[2])
	}

	r.storeName = r.storeCode.Name()

	return &r, nil
}
