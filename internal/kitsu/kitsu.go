package kitsu

import (
	"time"

	"github.com/MunifTanjim/stremthru/internal/cache"
	"github.com/MunifTanjim/stremthru/internal/oauth"
	"golang.org/x/oauth2"
)

var apiClientCache = cache.NewLRUCache[APIClient](&cache.CacheConfig{
	Lifetime: 1 * time.Hour,
	Name:     "kitsu:api-client",
})

func GetAPIClient(tokenId string) *APIClient {
	if tokenId == "" {
		panic("tokenId cannot be empty")
	}

	var cachedClient APIClient
	if apiClientCache.Get(tokenId, &cachedClient) {
		return &cachedClient
	}

	conf := APIClientConfig{}

	conf.OAuth = APIClientConfigOAuth{
		Config: oauth.KitsuOAuthConfig.Config,
		GetTokenSource: func(oauthConfig oauth2.Config) oauth2.TokenSource {
			otok, _ := oauth.GetOAuthTokenById(tokenId)
			if otok == nil {
				return nil
			}
			return oauth.DatabaseTokenSource(&oauth.DatabaseTokenSourceConfig{
				OAuth:             &oauth.KitsuOAuthConfig.Config,
				TokenSourceConfig: oauth.KitsuTokenSourceConfig,
			}, otok.ToToken())
		},
	}

	client := NewAPIClient(&conf)

	apiClientCache.Add(tokenId, *client)

	return client
}
