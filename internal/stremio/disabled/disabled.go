package stremio_disabled

import (
	"encoding/json"
	"net/http"

	"github.com/MunifTanjim/stremthru/internal/server"
	"github.com/MunifTanjim/stremthru/internal/shared"
)

func handleManifest(w http.ResponseWriter, r *http.Request) {
	if !shared.IsMethod(r, http.MethodGet) {
		shared.ErrorMethodNotAllowed(r).Send(w, r)
		return
	}

	manifest, err := GetDisabledManifest(r.PathValue("manifestUrl"))
	if err != nil {
		shared.SendError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	if err := json.NewEncoder(w).Encode(manifest); err != nil {
		LogError(r, "failed to encode json", err)
	}
}

func handleConfigure(w http.ResponseWriter, r *http.Request) {
	if !shared.IsMethod(r, http.MethodGet) {
		shared.ErrorMethodNotAllowed(r).Send(w, r)
		return
	}

	http.Redirect(w, r, "/stremio/sidekick/?addon_operation=manage&try_load_addons=1", http.StatusFound)
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := server.GetReqCtx(r)
		ctx.Log = log.With("request_id", ctx.RequestId)
		next.ServeHTTP(w, r)
		ctx.RedactURLPathValues(r, "manifestUrl")
	})
}

func AddStremioDisabledEndpoints(mux *http.ServeMux) {
	withCors := shared.Middleware(shared.EnableCORS)

	router := http.NewServeMux()

	router.HandleFunc("/{manifestUrl}/manifest.json", withCors(handleManifest))
	router.HandleFunc("/{manifestUrl}/configure", handleConfigure)

	mux.Handle("/stremio/disabled/", http.StripPrefix("/stremio/disabled", commonMiddleware(router)))
}
