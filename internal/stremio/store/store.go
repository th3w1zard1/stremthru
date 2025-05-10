package stremio_store

import (
	"net/http"
	"strings"
	"time"

	"github.com/MunifTanjim/stremthru/core"
	"github.com/MunifTanjim/stremthru/internal/config"
	"github.com/MunifTanjim/stremthru/internal/server"
	"github.com/MunifTanjim/stremthru/internal/shared"
	store_video "github.com/MunifTanjim/stremthru/internal/store/video"
	stremio_usenet "github.com/MunifTanjim/stremthru/internal/stremio/usenet"
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/stremio/store/configure", http.StatusFound)
}

func getContentType(r *http.Request) (string, *core.APIError) {
	contentType := r.PathValue("contentType")
	if contentType != ContentTypeOther {
		return "", shared.ErrorBadRequest(r, "unsupported type: "+contentType)
	}
	return contentType, nil
}

func handleAction(w http.ResponseWriter, r *http.Request) {
	if !IsMethod(r, http.MethodGet) {
		shared.ErrorMethodNotAllowed(r).Send(w, r)
		return
	}

	ud, err := getUserData(r)
	if err != nil {
		SendError(w, r, err)
		return
	}

	actionId := r.PathValue("actionId")
	idr, err := parseId(actionId)
	if err != nil {
		SendError(w, r, err)
		return
	}

	storeActionIdPrefix := getStoreActionIdPrefix(idr.getStoreCode())
	if !strings.HasPrefix(actionId, storeActionIdPrefix) {
		shared.ErrorBadRequest(r, "unsupported id: "+actionId).Send(w, r)
	}

	ctx, err := ud.GetRequestContext(r, idr)
	if err != nil || ctx.Store == nil {
		if err != nil {
			LogError(r, "failed to get request context", err)
		}
		store_video.Redirect("500", w, r)
		return
	}

	idPrefix := getIdPrefix(idr.getStoreCode())
	switch strings.TrimPrefix(actionId, storeActionIdPrefix) {
	case "clear_cache":
		cacheKey := getCatalogCacheKey(idPrefix, ctx.StoreAuthToken)
		catalogCache.Remove(cacheKey)
	}

	store_video.Redirect("200", w, r)
}

func handleStrem(w http.ResponseWriter, r *http.Request) {
	if !IsMethod(r, http.MethodGet) && !IsMethod(r, http.MethodHead) {
		shared.ErrorMethodNotAllowed(r).Send(w, r)
		return
	}

	ud, err := getUserData(r)
	if err != nil {
		SendError(w, r, err)
		return
	}

	videoIdWithLink := r.PathValue("videoId")
	idr, err := parseId(videoIdWithLink)
	if err != nil {
		SendError(w, r, err)
		return
	}

	idPrefix := getIdPrefix(idr.getStoreCode())
	if !strings.HasPrefix(videoIdWithLink, idPrefix) {
		shared.ErrorBadRequest(r, "unsupported id: "+videoIdWithLink).Send(w, r)
		return
	}

	ctx, err := ud.GetRequestContext(r, idr)
	if err != nil || ctx.Store == nil {
		if err != nil {
			LogError(r, "failed to get request context", err)
		}
		shared.ErrorBadRequest(r, "failed to get request context").Send(w, r)
		return
	}

	videoId := strings.TrimPrefix(videoIdWithLink, idPrefix)
	videoId, link, _ := strings.Cut(videoId, ":")

	url := link

	if url == "" {
		ctx.Log.Warn("no matching file found for (" + videoIdWithLink + ")")
		store_video.Redirect("no_matching_file", w, r)
		return
	}

	if idr.isUsenet {
		storeName := ctx.Store.GetName()
		rParams := &stremio_usenet.GenerateLinkParams{
			Link:     link,
			CLientIP: ctx.ClientIP,
		}
		rParams.APIKey = ctx.StoreAuthToken
		var lerr error
		data, err := stremio_usenet.GenerateLink(rParams, storeName)
		if err == nil {
			if config.StoreContentProxy.IsEnabled(string(storeName)) && ctx.StoreAuthToken == config.StoreAuthToken.GetToken(ctx.ProxyAuthUser, string(storeName)) {
				if ctx.IsProxyAuthorized {
					tunnelType := config.StoreTunnel.GetTypeForStream(string(ctx.Store.GetName()))
					if proxyLink, err := shared.CreateProxyLink(r, data.Link, nil, tunnelType, 12*time.Hour, ctx.ProxyAuthUser, ctx.ProxyAuthPassword, true, ""); err == nil {
						data.Link = proxyLink
					} else {
						lerr = err
					}
				}
			}
		} else {
			lerr = err
		}
		if lerr != nil {
			LogError(r, "failed to generate stremthru link", lerr)
			store_video.Redirect("500", w, r)
			return
		}

		http.Redirect(w, r, data.Link, http.StatusFound)
	} else {
		stLink, err := shared.GenerateStremThruLink(r, ctx, url)
		if err != nil {
			LogError(r, "failed to generate stremthru link", err)
			store_video.Redirect("500", w, r)
			return
		}

		http.Redirect(w, r, stLink.Link, http.StatusFound)
	}
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := server.GetReqCtx(r)
		ctx.Log = log.With("request_id", ctx.RequestId)
		next.ServeHTTP(w, r)
		ctx.RedactURLPathValues(r, "userData")
	})
}

func AddStremioStoreEndpoints(mux *http.ServeMux) {
	withCors := shared.Middleware(shared.EnableCORS)

	router := http.NewServeMux()

	router.HandleFunc("/{$}", handleRoot)

	router.HandleFunc("/manifest.json", withCors(handleManifest))
	router.HandleFunc("/{userData}/manifest.json", withCors(handleManifest))

	router.HandleFunc("/configure", handleConfigure)
	router.HandleFunc("/{userData}/configure", handleConfigure)

	router.HandleFunc("/{userData}/catalog/{contentType}/{idJson}", withCors(handleCatalog))
	router.HandleFunc("/{userData}/catalog/{contentType}/{id}/{extraJson}", withCors(handleCatalog))

	router.HandleFunc("/{userData}/meta/{contentType}/{idJson}", withCors(handleMeta))

	router.HandleFunc("/{userData}/stream/{contentType}/{idJson}", withCors(handleStream))

	router.HandleFunc("/{userData}/_/action/{actionId}", withCors(handleAction))
	router.HandleFunc("/{userData}/_/strem/{videoId}", withCors(handleStrem))

	mux.Handle("/stremio/store/", http.StripPrefix("/stremio/store", commonMiddleware(router)))
}
