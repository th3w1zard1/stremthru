package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
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

func ProxyHandler(w http.ResponseWriter, r *http.Request) {
	cred, hasCred := extractProxyAuthCred(r)
	if config.EnforceProxyAuth && (!hasCred || !config.ProxyAuthCredential[cred]) {
		w.Header().Add("Proxy-Authenticate", "Basic")
		http.Error(w, "proxy unauthorized", http.StatusProxyAuthRequired)
		return
	}

	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	targetUrl := r.URL.Query().Get("url")
	if targetUrl == "" {
		http.Error(w, "missing url", http.StatusBadRequest)
		return
	}

	targetUrl, err := url.QueryUnescape(targetUrl)
	if err != nil {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	if u, err := url.ParseRequestURI(targetUrl); err != nil || u.Scheme == "" || u.Host == "" {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	request, err := http.NewRequest(r.Method, targetUrl, nil)
	if err != nil {
		http.Error(w, "failed to create request", http.StatusInternalServerError)
		return
	}

	copyHeaders(r.Header, request.Header)

	response, err := httpClient.Do(request)
	if err != nil {
		http.Error(w, "failed to request url", http.StatusBadGateway)
		return
	}
	defer response.Body.Close()

	copyHeaders(response.Header, w.Header())

	w.WriteHeader(response.StatusCode)

	_, err = io.Copy(w, response.Body)
	if err != nil {
		log.Printf("stream failure: %v", err)
	}
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func main() {
	http.HandleFunc("/health", HealthHandler)
	http.HandleFunc("/proxy", ProxyHandler)

	addr := ":" + config.Port
	log.Println("stremthru listening on " + addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("failed to start stremthru: %v", err)
	}
}
