package torbox

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/MunifTanjim/stremthru/core"
)

type ResponseStatus string

const (
	ResponseStatusSuccess ResponseStatus = "success"
	ResponseStatusError   ResponseStatus = "error"
)

type ResponseError struct {
	Detail string    `json:"detail"`
	Err    ErrorCode `json:"error"`
	Data   string    `json:"data"`
}

func (e *ResponseError) Error() string {
	ret, _ := json.Marshal(e)
	return string(ret)
}

type Response[T any] struct {
	response[T]
	errData any `json:"-"`
}

type response[T any] struct {
	Success bool      `json:"success"`
	Data    T         `json:"data,omitempty"`
	Detail  string    `json:"detail"`
	Error   ErrorCode `json:"error,omitempty"`
}

func (r *Response[T]) UnmarshalJSON(data []byte) error {
	resp := response[T]{}
	respErr := json.Unmarshal(data, &resp)
	if respErr == nil {
		r.response = resp
		return nil
	}
	fallbackResp := response[any]{}
	if err := json.Unmarshal(data, &fallbackResp); err != nil {
		return err
	}
	if fallbackResp.Success {
		return respErr
	}
	r.Error = fallbackResp.Error
	r.Detail = fallbackResp.Detail
	r.errData = fallbackResp.Data
	return nil
}

type ResponseEnvelop interface {
	IsSuccess() bool
	GetError() *ResponseError
}

func (r Response[any]) IsSuccess() bool {
	return r.Success && r.Error == ""
}

func (r Response[any]) GetError() *ResponseError {
	if r.IsSuccess() {
		return nil
	}
	err := ResponseError{
		Err:    r.Error,
		Detail: r.Detail,
	}
	if data, ok := r.errData.(string); ok {
		err.Data = data
	}
	return &err
}

type APIResponse[T any] struct {
	Header     http.Header
	StatusCode int
	Data       T
	Detail     string
}

func newAPIResponse[T any](res *http.Response, data T, detail string) APIResponse[T] {
	apiResponse := APIResponse[T]{
		StatusCode: 503,
		Data:       data,
		Detail:     detail,
	}
	if res != nil {
		apiResponse.Header = res.Header
		apiResponse.StatusCode = res.StatusCode
	}
	return apiResponse
}

func extractResponseError(statusCode int, body []byte, v ResponseEnvelop) error {
	if !v.IsSuccess() {
		return v.GetError()
	}
	if statusCode >= http.StatusBadRequest {
		return errors.New(string(body))
	}
	return nil
}

func processResponseBody(res *http.Response, err error, v ResponseEnvelop) error {
	if err != nil {
		return err
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		return err
	}

	err = core.UnmarshalJSON(res.StatusCode, body, v)
	if err != nil {
		return err
	}

	return extractResponseError(res.StatusCode, body, v)
}
