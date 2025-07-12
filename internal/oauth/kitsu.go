package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MunifTanjim/stremthru/core"
	"github.com/MunifTanjim/stremthru/internal/config"
	"github.com/MunifTanjim/stremthru/internal/request"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type kitsuResponseError struct {
	Err     string `json:"error"`
	ErrDesc string `json:"error_description"`
}

func (e *kitsuResponseError) Error() string {
	ret, _ := json.Marshal(e)
	return string(ret)
}

func (e *kitsuResponseError) Unmarshal(res *http.Response, body []byte, v any) error {
	contentType := res.Header.Get("Content-Type")
	switch {
	case strings.Contains(contentType, "application/vnd.api+json"):
		return core.UnmarshalJSON(res.StatusCode, body, v)
	case strings.Contains(contentType, "text/html"):
		if res.StatusCode >= http.StatusBadRequest {
			errMsg := strings.TrimSpace(string(body))
			if errMsg == "" {
				errMsg = res.Status
			}
			return errors.New(errMsg)
		}
		fallthrough
	default:
		return fmt.Errorf("unexpected content type: %s", contentType)
	}
}

func (r *kitsuResponseError) GetError(res *http.Response) error {
	if r == nil || r.Err == "" {
		return nil
	}
	return r
}

var KitsuTokenSourceConfig = TokenSourceConfig{
	Provider: ProviderKitsu,
	GetUser: func(client *http.Client, oauthConfig *oauth2.Config) (userId, userName string, err error) {
		req, err := http.NewRequest("GET", "https://kitsu.io/api/edge/users?filter%5Bself%5D=true", nil)
		if err != nil {
			return "", "", err
		}
		req.Header.Set("Accept", "application/vnd.api+json")
		req.Header.Set("Content-Type", "application/vnd.api+json")
		res, err := client.Do(req)
		var response struct {
			kitsuResponseError
			Data []struct {
				Attributes struct {
					Email string `json:"email"`
					Name  string `json:"name"`
				} `json:"attributes"`
				Id string `json:"id"`
			} `json:"data"`
		}
		err = request.ProcessResponseBody(res, err, &response)
		if err != nil {
			return "", "", err
		}
		if len(response.Data) != 1 {
			return "", "", errors.New("failed to fetch user info")
		}
		return response.Data[0].Attributes.Email, response.Data[0].Attributes.Name, nil
	},
	PrepareToken: func(tok *oauth2.Token, id, userId string, userName string) *oauth2.Token {
		return tok.WithExtra(map[string]any{
			"id":         id,
			"provider":   ProviderKitsu,
			"user_id":    userId,
			"user_name":  userName,
			"scope":      tok.Extra("scope").(string),
			"created_at": time.Unix(int64(tok.Extra("created_at").(float64)), 0),
		})
	},
}

var kitsuOAuthConfig = oauth2.Config{
	ClientID:     config.Integration.Kitsu.ClientId,
	ClientSecret: config.Integration.Kitsu.ClientSecret,
	Endpoint: oauth2.Endpoint{
		TokenURL: "https://kitsu.io/api/oauth/token",
	},
}

var KitsuOAuthConfig = OAuthConfig{
	Config: kitsuOAuthConfig,
	PasswordCredentialsToken: func(username, password string) (*oauth2.Token, error) {
		tok, err := kitsuOAuthConfig.PasswordCredentialsToken(context.Background(), username, password)
		if err != nil {
			return nil, err
		}

		kitsuLog.Debug("fetching user info for new token")
		userId, userName, err := KitsuTokenSourceConfig.GetUser(
			oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(tok)),
			&kitsuOAuthConfig,
		)
		if err != nil {
			return nil, err
		}

		existingOTok, err := GetOAuthTokenByUserId(KitsuTokenSourceConfig.Provider, userId)
		if err != nil {
			return nil, err
		}

		if existingOTok != nil {
			client := oauth2.NewClient(
				context.Background(),
				DatabaseTokenSource(&DatabaseTokenSourceConfig{
					OAuth:             &kitsuOAuthConfig,
					TokenSourceConfig: KitsuTokenSourceConfig,
				}, existingOTok.ToToken()),
			)

			kitsuLog.Debug("fetching user info for existing token")
			uId, _, err := KitsuTokenSourceConfig.GetUser(
				client,
				&kitsuOAuthConfig,
			)
			if err != nil || uId != userId {
				existingOTok.AccessToken = ""
				existingOTok.RefreshToken = ""
				err = SaveOAuthToken(existingOTok)
				if err != nil {
					return nil, err
				}
				existingOTok = nil
			}
		}

		tokenId := uuid.NewString()
		if existingOTok != nil {
			tokenId = existingOTok.Id
		}

		tok = KitsuTokenSourceConfig.PrepareToken(tok, tokenId, userId, userName)

		otok := &OAuthToken{}
		otok = otok.FromToken(tok)
		err = SaveOAuthToken(otok)
		if err != nil {
			return nil, err
		}

		return tok, nil
	},
}
