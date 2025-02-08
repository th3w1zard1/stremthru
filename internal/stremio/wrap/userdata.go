package stremio_wrap

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/MunifTanjim/stremthru/core"
	"github.com/MunifTanjim/stremthru/internal/cache"
	"github.com/MunifTanjim/stremthru/internal/config"
	"github.com/MunifTanjim/stremthru/internal/context"
	"github.com/MunifTanjim/stremthru/internal/shared"
	"github.com/MunifTanjim/stremthru/internal/stremio/addon"
	"github.com/MunifTanjim/stremthru/stremio"
)

var upstreamResolverCache = cache.NewCache[upstreamsResolver](&cache.CacheConfig{
	Name:     "stremio:wrap:upstreamResolver",
	Lifetime: 24 * time.Hour,
})

type upstreamsResolverEntry struct {
	Prefix  string
	Indices []int
}

type upstreamsResolver map[string][]upstreamsResolverEntry

func (usr upstreamsResolver) resolve(ud UserData, rName stremio.ResourceName, rType string, id string) []UserDataUpstream {
	upstreams := []UserDataUpstream{}
	key := string(rName) + ":" + rType
	if _, found := usr[key]; !found {
		return upstreams
	}
	for _, entry := range usr[key] {
		if strings.HasPrefix(id, entry.Prefix) {
			for _, idx := range entry.Indices {
				upstreams = append(upstreams, ud.Upstreams[idx])
			}
			break
		}
	}
	return upstreams
}

type UserDataUpstream struct {
	URL            string                         `json:"u"`
	baseUrl        *url.URL                       `json:"-"`
	ExtractorId    string                         `json:"e,omitempty"`
	extractor      StreamTransformerExtractorBlob `json:"-"`
	NoContentProxy bool                           `json:"ncp,omitempty"`
}

type UserData struct {
	ManifestURL string             `json:"manifest_url,omitempty"`
	Upstreams   []UserDataUpstream `json:"upstreams"`
	StoreName   string             `json:"store"`
	StoreToken  string             `json:"token"`
	CachedOnly  bool               `json:"cached,omitempty"`

	TemplateId string                        `json:"template,omitempty"`
	template   StreamTransformerTemplateBlob `json:"-"`

	Sort string `json:"sort,omitempty"`

	encoded   string             `json:"-"`
	manifests []stremio.Manifest `json:"-"`
	resolver  upstreamsResolver  `json:"-"`
}

func (ud UserData) HasRequiredValues() bool {
	if len(ud.Upstreams) == 0 {
		return false
	}
	for i := range ud.Upstreams {
		if ud.Upstreams[i].URL == "" {
			return false
		}
	}
	return ud.StoreToken != ""
}

func (ud UserData) GetEncoded(forceRefresh bool) (string, error) {
	if ud.encoded == "" || forceRefresh {
		blob, err := json.Marshal(ud)
		if err != nil {
			return "", err
		}
		ud.encoded = core.Base64Encode(string(blob))
	}

	return ud.encoded, nil
}

type userDataError struct {
	upstreamUrl []string
	store       string
	token       string
}

func (uderr *userDataError) Error() string {
	var str strings.Builder
	hasSome := false
	for _, err := range uderr.upstreamUrl {
		if err != "" {
			str.WriteString("upstream_url: ")
			str.WriteString(err)
			hasSome = true
		}
	}
	if hasSome {
		str.WriteString(", ")
	}
	if uderr.store != "" {
		str.WriteString("store: ")
		str.WriteString(uderr.store)
	}
	if hasSome {
		str.WriteString(", ")
	}
	if uderr.token != "" {
		str.WriteString("token: ")
		str.WriteString(uderr.token)
	}
	return str.String()
}

func (ud UserData) GetRequestContext(r *http.Request) (*context.RequestContext, error) {
	ctx := &context.RequestContext{}

	upstreamUrlErrors := []string{}
	hasUpstreamUrlErrors := false
	for i := range ud.Upstreams {
		up := &ud.Upstreams[i]
		if up.baseUrl == nil {
			upstreamUrlErrors = append(upstreamUrlErrors, "Invalid Manifest URL("+up.URL+")")
			hasUpstreamUrlErrors = true
		} else {
			upstreamUrlErrors = append(upstreamUrlErrors, "")
		}
	}
	if hasUpstreamUrlErrors {
		return ctx, &userDataError{upstreamUrl: upstreamUrlErrors}
	}

	storeName := ud.StoreName
	storeToken := ud.StoreToken
	if storeName == "" && len(ud.Upstreams) > 0 {
		auth, err := core.ParseBasicAuth(storeToken)
		if err != nil {
			return ctx, &userDataError{token: err.Error()}
		}
		password := config.ProxyAuthPassword.GetPassword(auth.Username)
		if password != "" && password == auth.Password {
			ctx.IsProxyAuthorized = true
			ctx.ProxyAuthUser = auth.Username
			ctx.ProxyAuthPassword = auth.Password

			storeName = config.StoreAuthToken.GetPreferredStore(ctx.ProxyAuthUser)
			storeToken = config.StoreAuthToken.GetToken(ctx.ProxyAuthUser, storeName)
		}
	}

	if storeToken != "" {
		ctx.Store = shared.GetStore(storeName)
		ctx.StoreAuthToken = storeToken
	}

	ctx.ClientIP = shared.GetClientIP(r, ctx)

	return ctx, nil
}

func (ud UserData) getUpstreamManifests(ctx *context.RequestContext) ([]stremio.Manifest, []error) {
	if ud.manifests == nil {
		var wg sync.WaitGroup

		manifests := make([]stremio.Manifest, len(ud.Upstreams))
		errs := make([]error, len(ud.Upstreams))
		hasError := false
		for i := range ud.Upstreams {
			up := &ud.Upstreams[i]
			wg.Add(1)
			go func() {
				defer wg.Done()
				res, err := addon.GetManifest(&stremio_addon.GetManifestParams{BaseURL: up.baseUrl, ClientIP: ctx.ClientIP})
				manifests[i] = res.Data
				errs[i] = err
				if err != nil {
					hasError = true
				}
			}()
		}

		wg.Wait()

		if hasError {
			return manifests, errs
		}

		ud.manifests = manifests
	}

	return ud.manifests, nil
}

func (ud UserData) getUpstreamsResolver(ctx *context.RequestContext) (upstreamsResolver, error) {
	eud, err := ud.GetEncoded(false)
	if err != nil {
		return nil, err
	}
	if ud.resolver == nil {
		if upstreamResolverCache.Get(eud, &ud.resolver) {
			return ud.resolver, nil
		}

		manifests, errs := ud.getUpstreamManifests(ctx)
		if errs != nil {
			return nil, errors.Join(errs...)
		}

		resolver := upstreamsResolver{}
		entryIdxMap := map[string]int{}
		for mIdx := range manifests {
			m := &manifests[mIdx]
			for _, r := range m.Resources {
				if r.Name == stremio.ResourceNameAddonCatalog || r.Name == stremio.ResourceNameCatalog {
					continue
				}

				idPrefixes := getManifestResourceIdPrefixes(m, r)
				for _, rType := range getManifestResourceTypes(m, r) {
					key := string(r.Name) + ":" + string(rType)
					if _, found := resolver[key]; !found {
						resolver[key] = []upstreamsResolverEntry{}
					}
					for _, idPrefix := range idPrefixes {
						idPrefixKey := key + ":" + idPrefix
						if idx, found := entryIdxMap[idPrefixKey]; found {
							resolver[key][idx].Indices = append(resolver[key][idx].Indices, mIdx)
						} else {
							resolver[key] = append(resolver[key], upstreamsResolverEntry{
								Prefix:  idPrefix,
								Indices: []int{mIdx},
							})
							entryIdxMap[idPrefixKey] = len(resolver[key]) - 1
						}
					}
				}
			}
		}

		err = upstreamResolverCache.Add(eud, resolver)
		if err != nil {
			return nil, err
		}

		ud.resolver = resolver
	}

	return ud.resolver, nil
}

func (ud UserData) getUpstreams(ctx *context.RequestContext, rName stremio.ResourceName, rType, id string) ([]UserDataUpstream, error) {
	switch rName {
	case stremio.ResourceNameAddonCatalog:
		return []UserDataUpstream{}, nil
	case stremio.ResourceNameCatalog:
		return []UserDataUpstream{}, nil
	default:
		if len(ud.Upstreams) == 1 {
			return ud.Upstreams, nil
		}

		resolver, err := ud.getUpstreamsResolver(ctx)
		if err != nil {
			return nil, err
		}
		return resolver.resolve(ud, rName, rType, id), nil
	}
}

func getUserData(r *http.Request) (*UserData, error) {
	data := &UserData{}

	if IsMethod(r, http.MethodGet) || IsMethod(r, http.MethodHead) {
		data.encoded = r.PathValue("userData")
		if data.encoded == "" {
			return data, nil
		}
		blob, err := core.Base64DecodeToByte(data.encoded)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(blob, data)
		if err != nil {
			return nil, err
		}

		if data.ManifestURL != "" {
			data.Upstreams = []UserDataUpstream{
				{
					URL: data.ManifestURL,
				},
			}
			data.ManifestURL = ""
			_, err := data.GetEncoded(true)
			if err != nil {
				return nil, err
			}
		}

		hasMissingExtractor := false
		for i := range data.Upstreams {
			up := &data.Upstreams[i]
			if up.ExtractorId != "" {
				if err := extractorStore.Get(up.ExtractorId, &up.extractor); err != nil {
					core.LogError(fmt.Sprintf("[stremio/wrap] failed to fetch extractor(%s)", up.ExtractorId), err)
					hasMissingExtractor = true
				}
			} else {
				hasMissingExtractor = true
			}
		}

		if !hasMissingExtractor && data.TemplateId != "" {
			if err := templateStore.Get(data.TemplateId, &data.template); err != nil {
				core.LogError(fmt.Sprintf("[stremio/wrap] failed to fetch template(%s)", data.TemplateId), err)
			}
		}
	}

	if IsMethod(r, http.MethodPost) {
		err := r.ParseForm()
		if err != nil {
			return nil, err
		}

		upstreams_length := 1
		if v := r.Form.Get("upstreams_length"); v != "" {
			upstreams_length, err = strconv.Atoi(v)
			if err != nil {
				return nil, err
			}
		}

		data.Sort = r.Form.Get("sort")

		data.TemplateId = r.Form.Get("transformer.template_id")
		data.template = StreamTransformerTemplateBlob{
			Name:        r.Form.Get("transformer.template.name"),
			Description: r.Form.Get("transformer.template.description"),
		}

		for idx := range upstreams_length {
			upURL := r.Form.Get("upstreams[" + strconv.Itoa(idx) + "].url")
			if strings.HasPrefix(upURL, "stremio:") {
				upURL = "https:" + strings.TrimPrefix(upURL, "stremio:")
			}
			extractorId := r.Form.Get("upstreams[" + strconv.Itoa(idx) + "].transformer.extractor_id")
			up := UserDataUpstream{
				URL:         upURL,
				ExtractorId: extractorId,
			}
			extractor := r.Form.Get("upstreams[" + strconv.Itoa(idx) + "].transformer.extractor")
			if extractor != "" {
				up.extractor = StreamTransformerExtractorBlob(extractor)
			}
			if upURL != "" || extractorId != "" || extractor != "" {
				up.NoContentProxy = r.Form.Get("upstreams["+strconv.Itoa(idx)+"].no_content_proxy") == "on"
				data.Upstreams = append(data.Upstreams, up)
			}
		}

		data.StoreName = r.Form.Get("store")
		data.StoreToken = r.Form.Get("token")
		data.CachedOnly = r.Form.Get("cached") == "on"

		_, err = data.GetEncoded(false)
		if err != nil {
			return nil, err
		}
	}

	if !SupportAdvanced && len(data.Upstreams) > 1 {
		data.Upstreams = data.Upstreams[0:1]
	}

	for i := range data.Upstreams {
		up := &data.Upstreams[i]
		if up.URL != "" {
			if baseUrl, err := url.Parse(up.URL); err == nil {
				if strings.HasSuffix(baseUrl.Path, "/manifest.json") {
					baseUrl.Path = strings.TrimSuffix(baseUrl.Path, "/manifest.json")
					up.baseUrl = baseUrl
				}
			}
		}
	}

	return data, nil
}
