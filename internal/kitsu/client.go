package kitsu

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/MunifTanjim/stremthru/core"
	"github.com/MunifTanjim/stremthru/internal/config"
	"github.com/MunifTanjim/stremthru/internal/request"
	"golang.org/x/oauth2"
)

type APIClientConfigOAuth struct {
	Config         oauth2.Config
	GetTokenSource func(oauth2.Config) oauth2.TokenSource
}

type APIClientConfig struct {
	HTTPClient *http.Client
	OAuth      APIClientConfigOAuth
}

type APIClientOAuth struct {
	Config oauth2.Config
	client *APIClient
}

type APIClient struct {
	BaseURL    *url.URL
	httpClient *http.Client
	OAuth      APIClientOAuth

	reqQuery  func(query *url.Values, params request.Context)
	reqHeader func(query *http.Header, params request.Context)
}

func NewAPIClient(conf *APIClientConfig) *APIClient {
	if conf.HTTPClient == nil {
		conf.HTTPClient = config.DefaultHTTPClient
	}

	c := &APIClient{}

	baseUrl, err := url.Parse("https://kitsu.io/api/edge")
	if err != nil {
		panic(err)
	}

	c.BaseURL = baseUrl

	c.OAuth.Config = oauth2.Config{
		ClientID:     conf.OAuth.Config.ClientID,
		ClientSecret: conf.OAuth.Config.ClientSecret,
		Endpoint:     conf.OAuth.Config.Endpoint,
	}
	c.OAuth.client = c

	tokenSource := conf.OAuth.GetTokenSource(c.OAuth.Config)
	if tokenSource == nil {
		c.httpClient = conf.HTTPClient
	} else {
		c.httpClient = oauth2.NewClient(
			context.WithValue(context.Background(), oauth2.HTTPClient, conf.HTTPClient),
			tokenSource,
		)
	}

	c.reqQuery = func(query *url.Values, params request.Context) {
	}

	c.reqHeader = func(header *http.Header, params request.Context) {
		header.Set("Accept", "application/vnd.api+json")
		header.Set("Content-Type", "application/vnd.api+json")
	}

	return c
}

type Ctx = request.Ctx

type ResponseError struct {
	Err     string `json:"error,omitempty"`
	ErrDesc string `json:"error_description,omitempty"`
}

func (e *ResponseError) Error() string {
	ret, _ := json.Marshal(e)
	return string(ret)
}

type ResponseContainer interface {
	GetError(res *http.Response) error
	Unmarshal(res *http.Response, body []byte, v any) error
}

func (r *ResponseError) GetError(res *http.Response) error {
	if r == nil || r.Err == "" {
		return nil
	}
	return r
}

func (r *ResponseError) Unmarshal(res *http.Response, body []byte, v any) error {
	contentType := res.Header.Get("Content-Type")
	switch {
	case contentType == "application/vnd.api+json":
		return core.UnmarshalJSON(res.StatusCode, body, v)
	default:
		return errors.New("unexpected content type: " + contentType)
	}
}

func (c APIClient) Request(method, path string, params request.Context, v ResponseContainer) (*http.Response, error) {
	if params == nil {
		params = &Ctx{}
	}
	req, err := params.NewRequest(c.BaseURL, method, path, c.reqHeader, c.reqQuery)
	if err != nil {
		error := core.NewAPIError("failed to create request")
		error.Cause = err
		return nil, error
	}
	res, err := c.httpClient.Do(req)
	err = request.ProcessResponseBody(res, err, v)
	if err != nil {
		error := core.NewUpstreamError("")
		if rerr, ok := err.(*core.Error); ok {
			error.Msg = rerr.Msg
			error.Code = rerr.Code
			error.StatusCode = rerr.StatusCode
			error.UpstreamCause = rerr
		} else {
			error.Cause = err
		}
		error.InjectReq(req)
		return res, err
	}
	return res, nil
}

type APIResponse[T any] struct {
	Header     http.Header
	StatusCode int
	Data       T
}

func newAPIResponse[T any](res *http.Response, data T) APIResponse[T] {
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
