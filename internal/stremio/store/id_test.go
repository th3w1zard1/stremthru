package stremio_store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdParser(t *testing.T) {
	for _, tc := range []struct {
		name      string
		id        string
		idr       ParsedId
		storeCode string
	}{
		{"catalog rd", "st:store:rd", ParsedId{
			storeCode: "rd",
			storeName: "realdebrid",
		}, "rd"},
		{"catalog st-rd", "st:store:st-rd", ParsedId{
			isST:      true,
			storeCode: "rd",
			storeName: "realdebrid",
		}, "st-rd"},
		{"deprecated - catalog st", "st:store:st", ParsedId{
			isDeprecated: true,
			isST:         true,
		}, "st"},
		{"deprecated - catalog st:rd", "st:store:st:rd", ParsedId{
			isDeprecated: true,
			isST:         true,
			storeCode:    "rd",
			storeName:    "realdebrid",
		}, "st:rd"},
		{"meta rd", "st:store:rd:XXX", ParsedId{
			storeCode: "rd",
			storeName: "realdebrid",
		}, "rd"},
		{"meta st-rd", "st:store:st-rd:XXX", ParsedId{
			isST:      true,
			storeCode: "rd",
			storeName: "realdebrid",
		}, "st-rd"},
		{"deprecated - meta st", "st:store:st:XXX", ParsedId{
			isDeprecated: true,
			isST:         true,
		}, "st"},
		{"deprecated - meta st:rd", "st:store:st:rd:XXX", ParsedId{
			isDeprecated: true,
			isST:         true,
			storeCode:    "rd",
			storeName:    "realdebrid",
		}, "st:rd"},
		{"stream rd", "st:store:rd:XXX:scheme:%2F%2Fexample.com%2F000", ParsedId{
			storeCode: "rd",
			storeName: "realdebrid",
		}, "rd"},
		{"stream st-rd", "st:store:st-rd:XXX:scheme:%2F%2Fexample.com%2F000", ParsedId{
			isST:      true,
			storeCode: "rd",
			storeName: "realdebrid",
		}, "st-rd"},
		{"deprecated - stream st", "st:store:st:XXX:scheme:%2F%2Fexample.com%2F000", ParsedId{
			isDeprecated: true,
			isST:         true,
		}, "st"},
		{"deprecated - stream st:rd", "st:store:st:rd:XXX:scheme:%2F%2Fexample.com%2F000", ParsedId{
			isDeprecated: true,
			isST:         true,
			storeCode:    "rd",
			storeName:    "realdebrid",
		}, "st:rd"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			idr, err := parseId(tc.id)
			assert.Nil(t, err)
			assert.Equal(t, &tc.idr, idr)
			assert.Equal(t, tc.storeCode, idr.getStoreCode())
		})
	}
}
