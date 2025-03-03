package request

import (
	"net/http"
)

type APIResponse[T any] struct {
	Header     http.Header
	StatusCode int
	Data       T
}

func NewAPIResponse[T any](res *http.Response, data T) APIResponse[T] {
	apiResponse := APIResponse[T]{
		StatusCode: 503,
		Data:       data,
	}
	if res != nil {
		apiResponse.Header = res.Header
		apiResponse.StatusCode = res.StatusCode
	}
	return apiResponse
}
