package shared

import (
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/MunifTanjim/stremthru/core"
	"github.com/MunifTanjim/stremthru/internal/cache"
	"github.com/MunifTanjim/stremthru/internal/config"
	"github.com/MunifTanjim/stremthru/internal/context"
	"github.com/MunifTanjim/stremthru/store"
	"github.com/MunifTanjim/stremthru/store/alldebrid"
	"github.com/MunifTanjim/stremthru/store/debridlink"
	"github.com/MunifTanjim/stremthru/store/easydebrid"
	"github.com/MunifTanjim/stremthru/store/offcloud"
	"github.com/MunifTanjim/stremthru/store/pikpak"
	"github.com/MunifTanjim/stremthru/store/premiumize"
	"github.com/MunifTanjim/stremthru/store/realdebrid"
	"github.com/MunifTanjim/stremthru/store/torbox"
	"github.com/golang-jwt/jwt/v5"
)

var adStore = alldebrid.NewStoreClient(&alldebrid.StoreClientConfig{
	HTTPClient: config.GetHTTPClient(config.StoreTunnel.GetTypeForAPI("alldebrid")),
})
var dlStore = debridlink.NewStoreClient(&debridlink.StoreClientConfig{
	HTTPClient: config.GetHTTPClient(config.StoreTunnel.GetTypeForAPI("debridlink")),
})
var edStore = easydebrid.NewStoreClient(&easydebrid.StoreClientConfig{
	HTTPClient: config.GetHTTPClient(config.StoreTunnel.GetTypeForAPI("easydebrid")),
})
var pmStore = premiumize.NewStoreClient(&premiumize.StoreClientConfig{
	HTTPClient: config.GetHTTPClient(config.StoreTunnel.GetTypeForAPI("premiumize")),
})
var ppStore = pikpak.NewStoreClient(&pikpak.StoreClientConfig{
	HTTPClient: config.GetHTTPClient(config.StoreTunnel.GetTypeForAPI("pikpak")),
})
var ocStore = offcloud.NewStoreClient(&offcloud.StoreClientConfig{
	HTTPClient: config.GetHTTPClient(config.StoreTunnel.GetTypeForAPI("offcloud")),
})
var rdStore = realdebrid.NewStoreClient(&realdebrid.StoreClientConfig{
	HTTPClient: config.GetHTTPClient(config.StoreTunnel.GetTypeForAPI("realdebrid")),
	UserAgent:  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
})
var tbStore = torbox.NewStoreClient(&torbox.StoreClientConfig{
	HTTPClient: config.GetHTTPClient(config.StoreTunnel.GetTypeForAPI("torbox")),
})

func GetStore(name string) store.Store {
	switch store.StoreName(name) {
	case store.StoreNameAlldebrid:
		return adStore
	case store.StoreNameDebridLink:
		return dlStore
	case store.StoreNameEasyDebrid:
		return edStore
	case store.StoreNameOffcloud:
		return ocStore
	case store.StoreNamePikPak:
		return ppStore
	case store.StoreNamePremiumize:
		return pmStore
	case store.StoreNameRealDebrid:
		return rdStore
	case store.StoreNameTorBox:
		return tbStore
	default:
		return nil
	}
}

func GetStoreByCode(code string) store.Store {
	switch store.StoreCode(code) {
	case store.StoreCodeAllDebrid:
		return adStore
	case store.StoreCodeDebridLink:
		return dlStore
	case store.StoreCodeEasyDebrid:
		return edStore
	case store.StoreCodeOffcloud:
		return ocStore
	case store.StoreCodePikPak:
		return ppStore
	case store.StoreCodePremiumize:
		return pmStore
	case store.StoreCodeRealDebrid:
		return rdStore
	case store.StoreCodeTorBox:
		return tbStore
	default:
		return nil
	}
}

type proxyLinkTokenData struct {
	EncLink    string            `json:"enc_link"`
	EncFormat  string            `json:"enc_format"`
	TunnelType config.TunnelType `json:"tunt,omitempty"`
}

func CreateProxyLink(r *http.Request, link string, headers map[string]string, tunnelType config.TunnelType, expiresIn time.Duration, user, password string) (string, error) {
	linkBlob := link
	if headers != nil {
		for k, v := range headers {
			linkBlob += "\n" + k + ": " + v
		}
	}

	encryptedLink, err := core.Encrypt(password, linkBlob)
	if err != nil {
		return "", err
	}

	proxyLink := ExtractRequestBaseURL(r).JoinPath("/v0/proxy")

	claims := core.JWTClaims[proxyLinkTokenData]{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:  "stremthru",
			Subject: user,
		},
		Data: &proxyLinkTokenData{
			EncLink:    encryptedLink,
			EncFormat:  core.EncryptionFormat,
			TunnelType: tunnelType,
		},
	}
	if expiresIn != 0 {
		claims.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(expiresIn))
	}
	token, err := core.CreateJWT(password, claims)

	if err != nil {
		return "", err
	}

	proxyLink = proxyLink.JoinPath(token)

	if filename := filepath.Base(link); filename != "" && filepath.Ext(filename) != "" {
		proxyLink = proxyLink.JoinPath(filename)
	}

	return proxyLink.String(), nil
}

func GenerateStremThruLink(r *http.Request, ctx *context.StoreContext, link string) (*store.GenerateLinkData, error) {
	params := &store.GenerateLinkParams{}
	params.APIKey = ctx.StoreAuthToken
	params.Link = link
	if ctx.ClientIP != "" {
		params.ClientIP = ctx.ClientIP
	}

	data, err := ctx.Store.GenerateLink(params)
	if err != nil {
		return nil, err
	}

	storeName := string(ctx.Store.GetName())
	if config.StoreContentProxy.IsEnabled(storeName) && ctx.StoreAuthToken == config.StoreAuthToken.GetToken(ctx.ProxyAuthUser, storeName) {
		if ctx.IsProxyAuthorized {
			tunnelType := config.StoreTunnel.GetTypeForStream(string(ctx.Store.GetName()))
			proxyLink, err := CreateProxyLink(r, data.Link, nil, tunnelType, 12*time.Hour, ctx.ProxyAuthUser, ctx.ProxyAuthPassword)
			if err != nil {
				return nil, err
			}

			data.Link = proxyLink
		}
	}

	return data, nil
}

type proxyLink struct {
	User    string
	Value   string
	Headers map[string]string
	TunT    config.TunnelType
}

var proxyLinkTokenCache = func() cache.Cache[proxyLink] {
	return cache.NewCache[proxyLink](&cache.CacheConfig{
		Name:     "store:proxyLinkToken",
		Lifetime: 30 * time.Minute,
	})
}()

func getUserCredsFromJWT(t *jwt.Token) (user, password string, err error) {
	user, err = t.Claims.GetSubject()
	if err != nil {
		return "", "", err
	}
	password = config.ProxyAuthPassword.GetPassword(user)
	return user, password, nil
}

func UnwrapProxyLinkToken(encodedToken string) (user string, link string, headers map[string]string, tunnelType config.TunnelType, err error) {
	proxyLink := &proxyLink{}
	if found := proxyLinkTokenCache.Get(encodedToken, proxyLink); found {
		return proxyLink.User, proxyLink.Value, proxyLink.Headers, proxyLink.TunT, nil
	}

	claims := &core.JWTClaims[proxyLinkTokenData]{}
	password := ""
	_, err = core.ParseJWT(func(t *jwt.Token) (any, error) {
		user, password, err = getUserCredsFromJWT(t)
		return []byte(password), err
	}, encodedToken, claims)

	if err != nil {
		return "", "", nil, "", err
	}

	linkBlob, err := core.Decrypt(password, claims.Data.EncLink)
	if err != nil {
		return "", "", nil, "", err
	}

	link, headersBlob, hasHeaders := strings.Cut(linkBlob, "\n")

	proxyLink.User = user
	proxyLink.TunT = claims.Data.TunnelType
	proxyLink.Value = link

	if hasHeaders {
		proxyLink.Headers = map[string]string{}
		for _, header := range strings.Split(headersBlob, "\n") {
			if k, v, ok := strings.Cut(header, ": "); ok {
				proxyLink.Headers[k] = v
			}
		}
	}

	proxyLinkTokenCache.Add(encodedToken, *proxyLink)

	return proxyLink.User, proxyLink.Value, proxyLink.Headers, proxyLink.TunT, nil
}
