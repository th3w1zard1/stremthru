package stremio_root

import (
	"encoding/json"
	"net/http"

	"github.com/MunifTanjim/stremthru/internal/server"
	"github.com/MunifTanjim/stremthru/internal/shared"
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if !shared.IsMethod(r, http.MethodGet) {
		shared.ErrorMethodNotAllowed(r).Send(w, r)
		return
	}

	td := getTemplateData(r)

	page, err := getPage(td)
	if err != nil {
		shared.SendError(w, r, err)
		return
	}
	shared.SendHTML(w, 200, page)
	return
}

func handleManifest(w http.ResponseWriter, r *http.Request) {
	if !shared.IsMethod(r, http.MethodGet) {
		shared.ErrorMethodNotAllowed(r).Send(w, r)
		return
	}

	manifest := getManifest(r)

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

	http.Redirect(w, r, "/stremio/", http.StatusFound)
}

func handleAddonCatalog(w http.ResponseWriter, r *http.Request) {
	if !shared.IsMethod(r, http.MethodGet) {
		shared.ErrorMethodNotAllowed(r).Send(w, r)
		return
	}

	res := getAddonCatalog(r)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		LogError(r, "failed to encode json", err)
	}
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := server.GetReqCtx(r)
		ctx.Log = log.With("request_id", ctx.RequestId)
		next.ServeHTTP(w, r)
		ctx.RedactURLPathValues(r, "transportUrl")
	})
}

func AddStremioEndpoints(mux *http.ServeMux) {
	withCors := shared.Middleware(shared.EnableCORS)

	router := http.NewServeMux()

	router.HandleFunc("/{$}", handleRoot)
	router.HandleFunc("/manifest.json", withCors(handleManifest))
	router.HandleFunc("/addon_catalog/all/stremthru.json", withCors(handleAddonCatalog))
	router.HandleFunc("/configure", handleConfigure)

	mux.Handle("/stremio/", http.StripPrefix("/stremio", commonMiddleware(router)))
}
