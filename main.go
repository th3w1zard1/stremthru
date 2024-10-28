package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/MunifTanjim/stremthru/internal/config"
	"github.com/MunifTanjim/stremthru/internal/endpoint"
)

var httpClient = func() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DisableKeepAlives = true
	return &http.Client{
		Transport: transport,
	}
}()

func copyHeaders(src http.Header, dest http.Header) {
	for key, values := range src {
		for _, value := range values {
			dest.Add(key, value)
		}
	}
}

func extractProxyAuthCred(r *http.Request) (cred string, hasCred bool) {
	cred = r.Header.Get("Proxy-Authorization")
	if cred != "" {
		r.Header.Del("Proxy-Authorization")
		return strings.TrimPrefix(cred, "Basic "), true
	}
	cred = r.URL.Query().Get("token")
	return cred, cred != ""
}

func main() {
	mux := http.NewServeMux()

	endpoint.AddHealthEndpoints(mux)
	endpoint.AddProxyEndpoints(mux)
	endpoint.AddStoreEndpoints(mux)

	addr := ":" + config.Port
	server := &http.Server{Addr: addr, Handler: mux}

	log.Println("stremthru listening on " + addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("failed to start stremthru: %v", err)
	}
}
