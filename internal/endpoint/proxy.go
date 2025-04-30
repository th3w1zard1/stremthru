package endpoint

import (
	"net/http"

	"github.com/MunifTanjim/stremthru/core"
	"github.com/MunifTanjim/stremthru/internal/config"
	"github.com/MunifTanjim/stremthru/internal/server"
	"github.com/MunifTanjim/stremthru/internal/shared"
	store_video "github.com/MunifTanjim/stremthru/internal/store/video"
	"github.com/MunifTanjim/stremthru/internal/util"
)

func handleProxyLinkAccess(w http.ResponseWriter, r *http.Request) {
	ctx := server.GetReqCtx(r)
	ctx.RedactURLPathValues(r, "token")

	isGetReq := shared.IsMethod(r, http.MethodGet)
	if !isGetReq && !shared.IsMethod(r, http.MethodHead) {
		shared.ErrorMethodNotAllowed(r).Send(w, r)
		return
	}

	encodedToken := r.PathValue("token")
	if encodedToken == "" {
		shared.ErrorBadRequest(r, "missing token").Send(w, r)
		return
	}

	user, link, headers, tunnelType, err := shared.UnwrapProxyLinkToken(encodedToken)
	if err != nil {
		SendError(w, r, err)
		return
	}

	if headers != nil {
		for k, v := range headers {
			r.Header.Set(k, v)
		}
	}

	if isGetReq && user != "" {
		cpStore := contentProxyConnectionStore.WithScope(user)

		if limit := config.ContentProxyConnectionLimit.Get(user); limit > 0 {
			activeConnectionCount, err := cpStore.Count()
			if err != nil {
				ctx.Log.Error("[proxy] failed to count connections", "error", err)
			} else if activeConnectionCount >= limit {
				store_video.Redirect(store_video.StoreVideoNameContentProxyLimitReached, w, r)
				return
			}
		}

		if err := cpStore.Set(ctx.RequestId, contentProxyConnection{IP: core.GetRequestIP(r), Link: link}); err != nil {
			ctx.Log.Error("[proxy] failed to record connection", "error", err)
		} else {
			defer cpStore.Del(ctx.RequestId)
		}
	}
	bytesWritten, err := shared.ProxyResponse(w, r, link, tunnelType)
	ctx.Log.Info("[proxy] connection closed", "user", user, "size", util.ToSize(bytesWritten), "error", err)
}

func AddProxyEndpoints(mux *http.ServeMux) {
	withCors := shared.Middleware(shared.EnableCORS)

	mux.HandleFunc("/v0/proxy/{token}", withCors(handleProxyLinkAccess))
	mux.HandleFunc("/v0/proxy/{token}/{filename}", withCors(handleProxyLinkAccess))
}
