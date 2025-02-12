package stremio_wrap

import (
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/MunifTanjim/stremthru/core"
	"github.com/MunifTanjim/stremthru/internal/buddy"
	"github.com/MunifTanjim/stremthru/internal/cache"
	"github.com/MunifTanjim/stremthru/internal/config"
	"github.com/MunifTanjim/stremthru/internal/context"
	"github.com/MunifTanjim/stremthru/internal/server"
	"github.com/MunifTanjim/stremthru/internal/shared"
	"github.com/MunifTanjim/stremthru/internal/store/video"
	"github.com/MunifTanjim/stremthru/internal/stremio/addon"
	"github.com/MunifTanjim/stremthru/internal/stremio/configure"
	"github.com/MunifTanjim/stremthru/store"
	"github.com/MunifTanjim/stremthru/stremio"
	"golang.org/x/sync/singleflight"
)

var IsPublicInstance = config.IsPublicInstance
var MaxPublicInstanceUpstreamCount = 3

var addon = func() *stremio_addon.Client {
	return stremio_addon.NewClient(&stremio_addon.ClientConfig{})
}()

func parseCatalogId(id string, ud *UserData) (idx int, catalogId string, err error) {
	idxStr, catalogId, ok := strings.Cut(id, "::")
	if !ok {
		return -1, "", errors.New("invalid id")
	}
	idx, err = strconv.Atoi(idxStr)
	if err != nil {
		return -1, "", err
	}
	if len(ud.Upstreams) <= idx {
		return -1, "", errors.New("invalid id")
	}
	return idx, catalogId, nil
}

func (ud UserData) fetchAddonCatalog(ctx *context.StoreContext, w http.ResponseWriter, r *http.Request, rType, id string) {
	idx, catalogId, err := parseCatalogId(id, &ud)
	if err != nil {
		SendError(w, r, err)
		return
	}
	addon.ProxyResource(w, r, &stremio_addon.ProxyResourceParams{
		BaseURL:  ud.Upstreams[idx].baseUrl,
		Resource: string(stremio.ResourceNameAddonCatalog),
		Type:     rType,
		Id:       catalogId,
		ClientIP: ctx.ClientIP,
	})
}

func (ud UserData) fetchCatalog(ctx *context.StoreContext, w http.ResponseWriter, r *http.Request, rType, id, extra string) {
	idx, catalogId, err := parseCatalogId(id, &ud)
	if err != nil {
		SendError(w, r, err)
		return
	}
	addon.ProxyResource(w, r, &stremio_addon.ProxyResourceParams{
		BaseURL:  ud.Upstreams[idx].baseUrl,
		Resource: string(stremio.ResourceNameCatalog),
		Type:     rType,
		Id:       catalogId,
		Extra:    extra,
		ClientIP: ctx.ClientIP,
	})
}

func (ud UserData) fetchMeta(ctx *context.StoreContext, w http.ResponseWriter, r *http.Request, rType, id, extra string) error {
	upstreams, err := ud.getUpstreams(ctx, stremio.ResourceNameMeta, rType, id)
	if err != nil {
		return err
	}

	if len(upstreams) == 0 {
		shared.ErrorNotFound(r).Send(w, r)
		return nil
	}

	upstream := upstreams[0]

	addon.ProxyResource(w, r, &stremio_addon.ProxyResourceParams{
		BaseURL:  upstream.baseUrl,
		Resource: string(stremio.ResourceNameMeta),
		Type:     rType,
		Id:       id,
		Extra:    extra,
		ClientIP: ctx.ClientIP,
	})
	return nil
}

func (ud UserData) fetchStream(ctx *context.StoreContext, r *http.Request, rType, id string) (*stremio.StreamHandlerResponse, error) {
	log := ctx.Log

	eud, err := ud.GetEncoded(false)
	if err != nil {
		return nil, err
	}

	upstreams, err := ud.getUpstreams(ctx, stremio.ResourceNameStream, rType, id)
	if err != nil {
		return nil, err
	}
	upstreamsCount := len(upstreams)
	log.Debug("found addons for stream", "count", upstreamsCount)

	chunks := make([][]WrappedStream, upstreamsCount)
	errs := make([]error, upstreamsCount)

	template, err := ud.template.Parse()
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	for i := range upstreams {
		wg.Add(1)
		go func() {
			defer wg.Done()
			up := &upstreams[i]
			res, err := addon.FetchStream(&stremio_addon.FetchStreamParams{
				BaseURL:  up.baseUrl,
				Type:     rType,
				Id:       id,
				ClientIP: ctx.ClientIP,
			})
			streams := res.Data.Streams
			wstreams := make([]WrappedStream, len(streams))
			errs[i] = err
			if err == nil {
				extractor, err := up.extractor.Parse()
				if err != nil {
					errs[i] = err
				} else {
					transformer := StreamTransformer{
						Extractor: extractor,
						Template:  template,
					}
					for i := range streams {
						stream := &streams[i]
						wstream, err := transformer.Do(stream, up.ReconfigureStore)
						if err != nil {
							LogError(r, "failed to transform stream", err)
						}
						if up.NoContentProxy {
							wstream.noContentProxy = true
						}
						wstreams[i] = *wstream
					}
				}
			}
			chunks[i] = wstreams
		}()
	}
	wg.Wait()

	allStreams := []WrappedStream{}
	for i := range chunks {
		if errs[i] != nil {
			log.Error("failed to fetch streams", "error", errs[i])
			continue
		}
		allStreams = append(allStreams, chunks[i]...)
	}

	if template != nil {
		SortWrappedStreams(allStreams, ud.Sort)
	}

	totalStreams := len(allStreams)
	allStreams = dedupeStreams(allStreams)
	log.Debug("found streams", "total_count", totalStreams, "deduped_count", len(allStreams))

	hashes := []string{}
	magnetByHash := map[string]core.MagnetLink{}
	for i := range allStreams {
		stream := &allStreams[i]
		if stream.URL == "" && stream.InfoHash != "" {
			magnet, err := core.ParseMagnetLink(stream.InfoHash)
			if err != nil {
				continue
			}
			hashes = append(hashes, magnet.Hash)
			magnetByHash[magnet.Hash] = magnet
		}
	}

	stremId := strings.TrimSuffix(id, ".json")

	storeNamePrefix := ""
	isCachedByHash := map[string]bool{}
	if len(hashes) > 0 {
		cmParams := &store.CheckMagnetParams{Magnets: hashes}
		cmParams.APIKey = ctx.StoreAuthToken
		cmParams.ClientIP = ctx.ClientIP
		cmParams.SId = stremId
		cmRes, err := ctx.Store.CheckMagnet(cmParams)
		if err != nil {
			return nil, err
		}
		for _, item := range cmRes.Items {
			isCachedByHash[item.Hash] = item.Status == store.MagnetStatusCached
		}

		storeNamePrefix = "[" + strings.ToUpper(string(ctx.Store.GetName().Code())) + "] "
	}

	cachedStreams := []stremio.Stream{}
	uncachedStreams := []stremio.Stream{}
	for i := range allStreams {
		stream := &allStreams[i]
		if stream.URL == "" && stream.InfoHash != "" {
			magnet, ok := magnetByHash[strings.ToLower(stream.InfoHash)]
			if !ok {
				continue
			}
			stream.Name = storeNamePrefix + stream.Name
			surl := shared.ExtractRequestBaseURL(r).JoinPath("/stremio/wrap/" + eud + "/_/strem/" + magnet.Hash + "/" + strconv.Itoa(stream.FileIndex) + "/")
			if stream.BehaviorHints != nil && stream.BehaviorHints.Filename != "" {
				surl = surl.JoinPath(stream.BehaviorHints.Filename)
			}
			surl.RawQuery = "sid=" + stremId
			if stream.r != nil && stream.r.Season != "" && stream.r.Episode != "" {
				surl.RawQuery += "&re=" + url.QueryEscape(stream.r.Season+".{1,3}"+stream.r.Episode)
			}
			stream.URL = surl.String()
			stream.InfoHash = ""
			stream.FileIndex = 0

			if isCached, ok := isCachedByHash[magnet.Hash]; ok && isCached {
				stream.Name = "⚡ " + stream.Name
				cachedStreams = append(cachedStreams, *stream.Stream)
			} else if !ud.CachedOnly {
				uncachedStreams = append(uncachedStreams, *stream.Stream)
			}
		} else if stream.URL != "" {
			if !stream.noContentProxy {
				var headers map[string]string
				if stream.BehaviorHints != nil && stream.BehaviorHints.ProxyHeaders != nil && stream.BehaviorHints.ProxyHeaders.Request != nil {
					headers = stream.BehaviorHints.ProxyHeaders.Request
				}

				if url, err := shared.CreateProxyLink(r, ctx, stream.URL, headers, config.TUNNEL_TYPE_AUTO); err == nil && url != stream.URL {
					stream.URL = url
					stream.Name = "✨ " + stream.Name
				}
			}
			if stream.r == nil || stream.r.IsCached {
				cachedStreams = append(cachedStreams, *stream.Stream)
			} else {
				uncachedStreams = append(uncachedStreams, *stream.Stream)
			}
		}
	}

	streams := make([]stremio.Stream, len(cachedStreams)+len(uncachedStreams))
	idx := 0
	for i := range cachedStreams {
		streams[idx] = cachedStreams[i]
		idx++
	}
	for i := range uncachedStreams {
		streams[idx] = uncachedStreams[i]
		idx++
	}

	return &stremio.StreamHandlerResponse{
		Streams: streams,
	}, nil
}

func (ud UserData) fetchSubtitles(ctx *context.StoreContext, rType, id, extra string) (*stremio.SubtitlesHandlerResponse, error) {
	log := ctx.Log

	upstreams, err := ud.getUpstreams(ctx, stremio.ResourceNameSubtitles, rType, id)
	if err != nil {
		return nil, err
	}

	upstreamsCount := len(upstreams)
	log.Debug("found addons for subtitles", "count", upstreamsCount)

	chunks := make([][]stremio.Subtitle, upstreamsCount)
	errs := make([]error, len(upstreams))

	var wg sync.WaitGroup
	for i := range upstreams {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := addon.FetchSubtitles(&stremio_addon.FetchSubtitlesParams{
				BaseURL:  upstreams[i].baseUrl,
				Type:     rType,
				Id:       id,
				Extra:    extra,
				ClientIP: ctx.ClientIP,
			})
			chunks[i] = res.Data.Subtitles
			errs[i] = err
		}()
	}
	wg.Wait()

	subtitles := []stremio.Subtitle{}
	for i := range chunks {
		if errs[i] != nil {
			log.Error("failed to fetch subtitles", "error", errs[i])
			continue
		}
		subtitles = append(subtitles, chunks[i]...)
	}

	return &stremio.SubtitlesHandlerResponse{
		Subtitles: subtitles,
	}, nil
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/stremio/wrap/configure", http.StatusFound)
}

func handleManifest(w http.ResponseWriter, r *http.Request) {
	if !IsMethod(r, http.MethodGet) {
		shared.ErrorMethodNotAllowed(r).Send(w, r)
		return
	}

	ud, err := getUserData(r)
	if err != nil {
		SendError(w, r, err)
		return
	}

	ctx, err := ud.GetRequestContext(r)
	if err != nil {
		SendError(w, r, err)
		return
	}

	manifests, errs := ud.getUpstreamManifests(ctx)
	if errs != nil {
		serr := shared.ErrorInternalServerError(r, "failed to fetch upstream manifests")
		serr.Cause = errors.Join(errs...)
		serr.Send(w, r)
		return
	}

	manifest := getManifest(manifests, ud)

	SendResponse(w, r, 200, manifest)
}

func handleConfigure(w http.ResponseWriter, r *http.Request) {
	if !IsMethod(r, http.MethodGet) && !IsMethod(r, http.MethodPost) {
		shared.ErrorMethodNotAllowed(r).Send(w, r)
		return
	}

	ud, err := getUserData(r)
	if err != nil {
		SendError(w, r, err)
		return
	}

	td := getTemplateData(ud, w, r)
	for i := range td.Configs {
		conf := &td.Configs[i]
		switch conf.Key {
		case "store":
			conf.Default = ud.StoreName
		case "token":
			conf.Default = ud.StoreToken
		case "cached":
			if ud.CachedOnly {
				conf.Default = "checked"
			}
		}
	}

	if action := r.Header.Get("x-addon-configure-action"); action != "" {
		switch action {
		case "authorize":
			if !IsPublicInstance {
				user := r.Form.Get("user")
				pass := r.Form.Get("pass")
				if config.ProxyAuthPassword.GetPassword(user) == pass {
					setCookie(w, user, pass)
					td.IsAuthed = true
					if r.Header.Get("hx-request") == "true" {
						w.Header().Add("hx-refresh", "true")
					}
				}
			}
		case "deauthorize":
			unsetCookie(w)
			td.IsAuthed = false
		case "add-upstream":
			if td.IsAuthed || len(td.Upstreams) < MaxPublicInstanceUpstreamCount {
				td.Upstreams = append(td.Upstreams, UpstreamAddon{
					URL: "",
				})
			}
		case "remove-upstream":
			end := len(td.Upstreams) - 1
			if end == 0 {
				end = 1
			}
			td.Upstreams = append([]UpstreamAddon{}, td.Upstreams[0:end]...)
		case "set-extractor":
			idx, err := strconv.Atoi(r.Form.Get("upstream_index"))
			if err != nil {
				SendError(w, r, err)
				return
			}
			up := &td.Upstreams[idx]
			id := up.ExtractorId
			if id != "" {
				value, err := getExtractor(id)
				if err != nil {
					LogError(r, "failed to fetch extractor", err)
					up.ExtractorError = "Failed to fetch extractor"
				} else if _, err := value.Parse(); err != nil {
					LogError(r, "failed to parse extractor", err)
					up.ExtractorError = err.Error()
				}
				up.Extractor = value
			}
		case "save-extractor":
			if td.IsAuthed {
				id := r.Form.Get("extractor_id")
				idx, err := strconv.Atoi(r.Form.Get("extractor_upstream_index"))
				if err != nil {
					SendError(w, r, err)
					return
				}
				up := &td.Upstreams[idx]
				value := up.Extractor
				if strings.HasPrefix(id, BUILTIN_TRANSFORMER_ENTITY_ID_EMOJI) {
					up.ExtractorError = "✨-prefixed ids are reserved"
				}
				if up.ExtractorError == "" {
					if _, err := value.Parse(); err != nil {
						LogError(r, "failed to parse extractor", err)
						up.ExtractorError = err.Error()
					} else {
						up.ExtractorId = id
						up.Extractor = value
					}
				}
				if up.ExtractorError == "" {
					if value == "" {
						if err := extractorStore.Del(id); err != nil {
							LogError(r, "failed to delete extractor", err)
							up.ExtractorError = "Failed to delete extractor"
						}
						extractorIds := []string{}
						for _, extractorId := range td.ExtractorIds {
							if extractorId != id {
								extractorIds = append(extractorIds, extractorId)
							}
						}
						td.ExtractorIds = extractorIds
						for i := range td.Upstreams {
							up := &td.Upstreams[i]
							if up.ExtractorId == id {
								up.ExtractorId = ""
								up.Extractor = ""
							}
						}
					} else {
						if err := extractorStore.Set(id, value); err != nil {
							LogError(r, "failed to save extractor", err)
							up.ExtractorError = "Failed to save extractor"
						} else {
							extractorIds := []string{}
							for _, extractorId := range td.ExtractorIds {
								if extractorId != id {
									extractorIds = append(extractorIds, extractorId)
								}
							}
							extractorIds = append(extractorIds, id)
							td.ExtractorIds = extractorIds
						}
					}
				}
			}
		case "set-template":
			id := ud.TemplateId
			if id != "" {
				value, err := getTemplate(id)
				if err != nil {
					LogError(r, "failed to fetch template", err)
					td.TemplateError.Name = "Failed to fetch template"
					td.TemplateError.Description = "Failed to fetch template"
				} else if t, err := value.Parse(); err != nil {
					LogError(r, "failed to parse template", err)
					if t.Name == nil {
						td.TemplateError.Name = err.Error()
					} else {
						td.TemplateError.Description = err.Error()
					}
				}
				td.Template = value
			}
		case "save-template":
			if td.IsAuthed {
				id := r.Form.Get("template_id")
				value := td.Template
				if strings.HasPrefix(id, BUILTIN_TRANSFORMER_ENTITY_ID_EMOJI) {
					td.TemplateError.Name = "✨-prefixed ids are reserved"
					td.TemplateError.Description = "✨-prefixed ids are reserved"
				}
				if td.TemplateError.IsEmpty() {
					if t, err := value.Parse(); err != nil {
						LogError(r, "failed to parse template", err)
						if t.Name == nil {
							td.TemplateError.Name = err.Error()
						} else {
							td.TemplateError.Description = err.Error()
						}
					} else {
						td.TemplateId = id
					}
				}
				if td.TemplateError.IsEmpty() {
					if value.Name == "" && value.Description == "" {
						if err := templateStore.Del(id); err != nil {
							LogError(r, "failed to delete template", err)
							td.TemplateError.Name = "Failed to delete template"
							td.TemplateError.Description = "Failed to delete template"
						}
						templateIds := []string{}
						for _, templateId := range td.TemplateIds {
							if templateId != id {
								templateIds = append(templateIds, templateId)
							}
						}
						td.TemplateIds = templateIds
						td.TemplateId = ""
						td.Template = StreamTransformerTemplateBlob{}
					} else {
						if err := templateStore.Set(id, value); err != nil {
							LogError(r, "failed to save template", err)
							td.TemplateError.Name = "Failed to save template"
							td.TemplateError.Description = "Failed to save template"
						} else {
							templateIds := []string{}
							for _, templateId := range td.TemplateIds {
								if templateId != id {
									templateIds = append(templateIds, templateId)
								}
							}
							templateIds = append(templateIds, id)
							td.TemplateIds = templateIds
						}
					}
				}
			}
		}

		page, err := getPage(td)
		if err != nil {
			SendError(w, r, err)
			return
		}
		SendHTML(w, 200, page)
		return
	}

	if ud.encoded != "" {
		var store_config *configure.Config
		var token_config *configure.Config
		for i := range td.Configs {
			conf := &td.Configs[i]
			switch conf.Key {
			case "store":
				store_config = conf
			case "token":
				token_config = conf
			}
		}

		ctx, err := ud.GetRequestContext(r)
		if err != nil {
			if uderr, ok := err.(*userDataError); ok {
				for i, err := range uderr.upstreamUrl {
					td.Upstreams[i].Error = err
				}
				store_config.Error = uderr.store
				token_config.Error = uderr.token
			} else {
				SendError(w, r, err)
				return
			}
		}

		if !td.HasUpstreamError() {
			manifests, errs := ud.getUpstreamManifests(ctx)
			for i := range manifests {
				tup := &td.Upstreams[i]
				manifest := manifests[i]

				if tup.Error == "" {
					if errs != nil && errs[i] != nil {
						LogError(r, "failed to fetch manifest", errs[i])
						tup.Error = "Failed to fetch Manifest"
						continue
					}

					if manifest.BehaviorHints != nil && manifest.BehaviorHints.Configurable {
						tup.IsConfigurable = true
						if manifest.BehaviorHints.ConfigurationRequired {
							tup.Error = "Configuration Required"
							continue
						}
					}
				}
			}

			if ctx.Store == nil {
				if ud.StoreName == "" {
					token_config.Error = "Invalid Token"
				} else {
					store_config.Error = "Invalid Store"
				}
			} else {
				params := &store.GetUserParams{}
				params.APIKey = ctx.StoreAuthToken
				_, err := ctx.Store.GetUser(params)
				if err != nil {
					LogError(r, "failed to access store", err)
					token_config.Error = "Failed to access store"
				}
			}
		}
	}

	hasError := td.HasFieldError()

	if IsMethod(r, http.MethodGet) || hasError {
		if !hasError && ud.HasRequiredValues() {
			if eud, err := ud.GetEncoded(false); err == nil {
				td.ManifestURL = ExtractRequestBaseURL(r).JoinPath("/stremio/wrap/" + eud + "/manifest.json").String()
			}
		}

		page, err := getPage(td)
		if err != nil {
			SendError(w, r, err)
			return
		}
		SendHTML(w, 200, page)
		return
	}

	eud, err := ud.GetEncoded(true)
	if err != nil {
		SendError(w, r, err)
		return
	}

	url := ExtractRequestBaseURL(r).JoinPath("/stremio/wrap/" + eud + "/configure")
	q := url.Query()
	q.Set("try_install", "1")
	url.RawQuery = q.Encode()

	http.Redirect(w, r, url.String(), http.StatusFound)
}

func handleResource(w http.ResponseWriter, r *http.Request) {
	if !IsMethod(r, http.MethodGet) && !IsMethod(r, http.MethodHead) {
		shared.ErrorMethodNotAllowed(r).Send(w, r)
		return
	}

	ud, err := getUserData(r)
	if err != nil {
		SendError(w, r, err)
		return
	}

	resource := r.PathValue("resource")
	contentType := r.PathValue("contentType")
	id := r.PathValue("id")
	extra := r.PathValue("extra")

	ctx, err := ud.GetRequestContext(r)
	if err != nil {
		SendError(w, r, err)
		return
	}

	switch stremio.ResourceName(resource) {
	case stremio.ResourceNameAddonCatalog:
		ud.fetchAddonCatalog(ctx, w, r, contentType, id)
	case stremio.ResourceNameCatalog:
		ud.fetchCatalog(ctx, w, r, contentType, id, extra)
	case stremio.ResourceNameMeta:
		err = ud.fetchMeta(ctx, w, r, contentType, id, extra)
		if err != nil {
			SendError(w, r, err)
		}
		return
	case stremio.ResourceNameStream:
		res, err := ud.fetchStream(ctx, r, contentType, id)
		if err != nil {
			SendError(w, r, err)
			return
		}
		SendResponse(w, r, 200, res)
		return

	case stremio.ResourceNameSubtitles:
		res, err := ud.fetchSubtitles(ctx, contentType, id, extra)
		if err != nil {
			SendError(w, r, err)
			return
		}
		SendResponse(w, r, 200, res)
		return
	default:
		addon.ProxyResource(w, r, &stremio_addon.ProxyResourceParams{
			BaseURL:  ud.Upstreams[0].baseUrl,
			Resource: resource,
			Type:     contentType,
			Id:       id,
			Extra:    extra,
			ClientIP: ctx.ClientIP,
		})
	}
}

func waitForMagnetStatus(ctx *context.StoreContext, m *store.GetMagnetData, status store.MagnetStatus, maxRetry int, retryInterval time.Duration) (*store.GetMagnetData, error) {
	retry := 0
	for m.Status != status && retry < maxRetry {
		gmParams := &store.GetMagnetParams{Id: m.Id}
		gmParams.APIKey = ctx.StoreAuthToken
		magnet, err := ctx.Store.GetMagnet(gmParams)
		if err != nil {
			return m, err
		}
		m = magnet
		time.Sleep(retryInterval)
		retry++
	}
	if m.Status != status {
		error := core.NewStoreError("torrent failed to reach status: " + string(status))
		error.StoreName = string(ctx.Store.GetName())
		return m, error
	}
	return m, nil
}

var stremLinkCache = cache.NewCache[string](&cache.CacheConfig{
	Name:     "stremio:wrap:streamLink",
	Lifetime: 3 * time.Hour,
})

func redirectToStaticVideo(w http.ResponseWriter, r *http.Request, cacheKey string, videoName string) {
	url := store_video.Redirect(videoName, w, r)
	stremLinkCache.AddWithLifetime(cacheKey, url, 1*time.Minute)
}

var stremGroup singleflight.Group

type stremResult struct {
	link        string
	error_log   string
	error_video string
}

func handleStrem(w http.ResponseWriter, r *http.Request) {
	if !IsMethod(r, http.MethodGet) && !IsMethod(r, http.MethodHead) {
		shared.ErrorMethodNotAllowed(r).Send(w, r)
		return
	}

	log := server.GetReqCtx(r).Log

	magnetHash := r.PathValue("magnetHash")
	fileName := r.PathValue("fileName")
	fileIdx := -1
	if idx, err := strconv.Atoi(r.PathValue("fileIdx")); err == nil {
		fileIdx = idx
	}

	ud, err := getUserData(r)
	if err != nil {
		SendError(w, r, err)
		return
	}

	ctx, err := ud.GetRequestContext(r)
	if err != nil || ctx.Store == nil {
		if err != nil {
			LogError(r, "failed to get request context", err)
		}
		shared.ErrorBadRequest(r, "failed to get request context").Send(w, r)
		return
	}

	cacheKey := strings.Join([]string{ctx.ClientIP, string(ctx.Store.GetName()), ctx.StoreAuthToken, magnetHash, strconv.Itoa(fileIdx), fileName}, ":")

	stremLink := ""
	if stremLinkCache.Get(cacheKey, &stremLink) {
		log.Debug("redirecting to cached stream link")
		http.Redirect(w, r, stremLink, http.StatusFound)
		return
	}

	result, err, _ := stremGroup.Do(cacheKey, func() (interface{}, error) {
		log.Debug("creating stream link")
		amParams := &store.AddMagnetParams{
			Magnet:   magnetHash,
			ClientIP: ctx.ClientIP,
		}
		amParams.APIKey = ctx.StoreAuthToken
		amRes, err := ctx.Store.AddMagnet(amParams)
		if err != nil {
			return &stremResult{
				error_log:   "failed to add magnet",
				error_video: "download_failed",
			}, err
		}

		magnet := &store.GetMagnetData{
			Id:      amRes.Id,
			Name:    amRes.Name,
			Hash:    amRes.Hash,
			Status:  amRes.Status,
			Files:   amRes.Files,
			AddedAt: amRes.AddedAt,
		}

		magnet, err = waitForMagnetStatus(ctx, magnet, store.MagnetStatusDownloaded, 3, 5*time.Second)

		query := r.URL.Query()
		sid := query.Get("sid")
		if sid == "" {
			sid = "*"
		}
		var pattern *regexp.Regexp
		if re := query.Get("re"); re != "" {
			if pat, err := regexp.Compile(re); err == nil {
				pattern = pat
			}
		}

		go buddy.TrackMagnet(ctx.Store, magnet.Hash, magnet.Files, sid, magnet.Status != store.MagnetStatusDownloaded, ctx.StoreAuthToken)

		if err != nil {
			strem := &stremResult{
				error_log:   "failed wait for magnet status",
				error_video: "500",
			}
			if magnet.Status == store.MagnetStatusQueued || magnet.Status == store.MagnetStatusDownloading || magnet.Status == store.MagnetStatusProcessing {
				strem.error_video = "downloading"
			} else if magnet.Status == store.MagnetStatusFailed || magnet.Status == store.MagnetStatusInvalid || magnet.Status == store.MagnetStatusUnknown {
				strem.error_video = "download_failed"
			}
			return strem, err
		}

		var file *store.MagnetFile
		if fileName != "" {
			for i := range magnet.Files {
				f := &magnet.Files[i]
				if f.Name == fileName {
					file = f
					log.Debug("matched file using filename", "filename", f.Name)
					break
				}
			}
		}
		if file == nil && pattern != nil {
			for i := range magnet.Files {
				f := &magnet.Files[i]
				if pattern.MatchString(f.Name) {
					file = f
					log.Debug("matched file using pattern", "pattern", pattern.String(), "filename", f.Name)
					break
				}
			}
		}
		if file == nil && fileIdx != -1 {
			for i := range magnet.Files {
				f := &magnet.Files[i]
				if f.Idx == fileIdx {
					log.Debug("matched file using fileidx", "fileidx", f.Idx, "filename", f.Name)
					file = f
					break
				}
			}
		}
		if file == nil {
			for i := range magnet.Files {
				f := &magnet.Files[i]
				if file == nil || file.Size < f.Size {
					log.Debug("matched file using largest size", "filename", f.Name)
					file = f
				}
			}
		}

		link := ""
		if file != nil {
			link = file.Link
		}
		if link == "" {
			return &stremResult{
				error_log:   "no matching file found for (" + sid + " - " + magnet.Hash + ")",
				error_video: "no_matching_file",
			}, nil
		}

		glRes, err := shared.GenerateStremThruLink(r, ctx, link)
		if err != nil {
			return &stremResult{
				error_log:   "failed to generate stremthru link",
				error_video: "500",
			}, err
		}

		stremLinkCache.Add(cacheKey, glRes.Link)

		return &stremResult{
			link: glRes.Link,
		}, nil
	})

	strem := result.(*stremResult)

	if strem.error_log != "" {
		if err != nil {
			LogError(r, strem.error_log, err)
		} else {
			log.Error(strem.error_log)
		}
		redirectToStaticVideo(w, r, cacheKey, strem.error_video)
		return
	}

	log.Debug("redirecting to stream link")
	http.Redirect(w, r, strem.link, http.StatusFound)
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := server.GetReqCtx(r)
		ctx.Log = log.With("request_id", ctx.RequestId)
		next.ServeHTTP(w, r)
		ctx.RedactURLPathValues(r, "userData")
	})
}

func AddStremioWrapEndpoints(mux *http.ServeMux) {
	seedDefaultTransformerEntities()

	withCors := shared.Middleware(shared.EnableCORS)

	router := http.NewServeMux()

	router.HandleFunc("/{$}", handleRoot)

	router.HandleFunc("/manifest.json", withCors(handleManifest))
	router.HandleFunc("/{userData}/manifest.json", withCors(handleManifest))

	router.HandleFunc("/configure", handleConfigure)
	router.HandleFunc("/{userData}/configure", handleConfigure)

	router.HandleFunc("/{userData}/{resource}/{contentType}/{id}", withCors(handleResource))
	router.HandleFunc("/{userData}/{resource}/{contentType}/{id}/{extra}", withCors(handleResource))

	router.HandleFunc("/{userData}/_/strem/{magnetHash}/{fileIdx}/{$}", withCors(handleStrem))
	router.HandleFunc("/{userData}/_/strem/{magnetHash}/{fileIdx}/{fileName}", withCors(handleStrem))

	mux.Handle("/stremio/wrap/", http.StripPrefix("/stremio/wrap", commonMiddleware(router)))
}
