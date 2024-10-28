package store

import "github.com/MunifTanjim/stremthru/core"

type ErrorCode = core.ErrorCode

const (
	ErrorCodeInvalidStoreName ErrorCode = "INVALID_STORE_NAME"
)

var ErrorInvalidStoreName = func(name string) *core.StoreError {
	err := core.NewStoreError("invalid store name")
	err.Code = ErrorCodeInvalidStoreName
	err.StoreName = name
	return err
}
