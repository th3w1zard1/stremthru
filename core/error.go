package core

import (
	"encoding/json"
	"net/http"
	"strings"
)

type ErrorType string

const (
	ErrorTypeAPI      ErrorType = "api_error"
	ErrorTypeStore    ErrorType = "store_error"
	ErrorTypeUpstream ErrorType = "upstream_error"
	ErrorTypeUnknown  ErrorType = "unknown_error"
)

type ErrorCode string

const (
	ErrorCodeBadRequest         ErrorCode = "BAD_REQUEST"
	ErrorCodeMagnetInvalidId    ErrorCode = "MAGNET_INVALID_ID"
	ErrorCodeMagnetInvalidURI   ErrorCode = "MAGNET_INVALID_URI"
	ErrorCodeStoreLimitExceeded ErrorCode = "STORE_LIMIT_EXCEEDED"
)

type StremThruError interface {
	Pack()
	GetStatusCode() int
	GetError() *Error
}

type Error struct {
	Type ErrorType `json:"type"`

	Code ErrorCode `json:"code,omitempty"`
	Msg  string    `json:"message"`

	Method     string `json:"method,omitempty"`
	Path       string `json:"path,omitempty"`
	StatusCode int    `json:"status_code,omitempty"`

	StoreName     string `json:"store_name,omitempty"`
	UpstreamCause error  `json:"__upstream_cause__,omitempty"`

	Cause error `json:"__cause__,omitempty"`
}

func (e *Error) Error() string {
	ret, _ := json.Marshal(e)
	return string(ret)
}

func (e *Error) Unwrap() error {
	return e.Cause
}

func (e *Error) InjectReq(r *http.Request) {
	e.Method = r.Method
	e.Path = r.URL.Path
	if storeName := r.Header.Get("X-StremThru-Store-Name"); storeName != "" {
		e.StoreName = storeName
	}
}

func (e *Error) Pack() {
	if e.StatusCode == 0 {
		e.StatusCode = 500
	}
	if e.Msg == "" {
		if e.Cause != nil {
			e.Msg = e.Cause.Error()
		} else if e.UpstreamCause != nil {
			e.Msg = e.UpstreamCause.Error()
		}
	}
}

func (e *Error) GetStatusCode() int {
	return e.StatusCode
}

func (e *Error) GetError() *Error {
	return e
}

func NewError(msg string) *Error {
	err := &Error{}
	err.Type = ErrorTypeUnknown
	err.Msg = msg
	return err
}

type err = Error

type APIError struct {
	err
}

func NewAPIError(msg string) *APIError {
	err := &APIError{}
	err.Type = ErrorTypeAPI
	err.Msg = msg
	return err
}

type StoreError struct {
	err
}

func NewStoreError(msg string) *StoreError {
	err := &StoreError{}
	err.Type = ErrorTypeStore
	err.Msg = msg
	return err
}

type UpstreamError struct {
	err
}

func (e *UpstreamError) Pack() {
	e.err.Pack()

	if e.StoreName != "" && e.Code != "" {
		e.Code = ErrorCode(strings.ToUpper(e.StoreName) + "_" + string(e.Code))
	}
}

func NewUpstreamError(msg string) *UpstreamError {
	err := &UpstreamError{}
	err.Type = ErrorTypeUpstream
	err.Msg = msg
	return err
}
