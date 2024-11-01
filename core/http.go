package core

import "net/http"

var DefaultHTTPTransport = func() *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DisableKeepAlives = true
	return transport
}()

var DefaultHTTPClient = func() *http.Client {
	return &http.Client{
		Transport: DefaultHTTPTransport,
	}
}()
