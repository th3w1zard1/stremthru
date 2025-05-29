package request

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/MunifTanjim/stremthru/core"
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

type ResponseContainer interface {
	GetError() error
}

func ProcessResponseBody(res *http.Response, err error, v ResponseContainer) error {
	if err != nil {
		return err
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		return err
	}

	contentType := res.Header.Get("Content-Type")

	switch {
	case strings.Contains(contentType, "application/json"):
		err := core.UnmarshalJSON(res.StatusCode, body, v)
		if err != nil {
			return err
		}
		return v.GetError()
	case strings.Contains(contentType, "text/html") && res.StatusCode >= http.StatusBadRequest:
		return errors.New(string(body))
	default:
		return errors.New("unsupported content-type: " + contentType)
	}
}
