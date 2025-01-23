package config

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type TunnelMap map[string]url.URL

func (tm TunnelMap) getProxy(hostname string) *url.URL {
	if proxy, ok := tm[hostname]; ok {
		return &proxy
	}
	return nil
}

// If tunnel is configured for `hostname` use that.
// Otherwise fallback to environment proxy, i.e. `HTTP_PROXY`, `HTTPS_PROXY`, `NO_PROXY`
func (tm TunnelMap) hostnameProxy(r *http.Request) (*url.URL, error) {
	proxy := tm.getProxy(r.URL.Hostname())
	if proxy == nil {
		return http.ProxyFromEnvironment(r)
	}
	if proxy.Host == "" {
		return nil, nil
	}
	return proxy, nil
}

// Ignores NO_PROXY
func (tm TunnelMap) forcedProxy(r *http.Request) (*url.URL, error) {
	if proxy := tm.getProxy("*"); proxy != nil && proxy.Host != "" {
		return proxy, nil
	}
	return nil, nil
}

func (tm TunnelMap) GetProxy(proxyType string) func(req *http.Request) (*url.URL, error) {
	switch proxyType {
	case "hostname":
		return tm.hostnameProxy
	case "forced":
		return tm.forcedProxy
	case "none":
		return nil
	default:
		panic("invalid proxy type")
	}
}

var Tunnel = func() TunnelMap {
	tunnelMap := make(TunnelMap)

	defaultProxy := &url.URL{}

	if value := getEnv("STREMTHRU_HTTP_PROXY", ""); len(value) > 0 {
		if err := os.Setenv("HTTP_PROXY", value); err != nil {
			log.Fatal("failed to set http_proxy")
		}
		if err := os.Setenv("HTTPS_PROXY", value); err != nil {
			log.Fatal("failed to set https_proxy")
		}
		if u, err := url.Parse(value); err == nil {
			defaultProxy = u
		}
	}

	// deprecated
	if value := getEnv("STREMTHRU_HTTPS_PROXY", getEnv("STREMTHRU_HTTP_PROXY", "")); len(value) > 0 {
		if err := os.Setenv("HTTPS_PROXY", value); err != nil {
			log.Fatal("failed to set https_proxy")
		}
		if defaultProxy.Host == "" {
			if u, err := url.Parse(value); err == nil {
				defaultProxy = u
			}
		}
	}

	tunnelMap["*"] = *defaultProxy

	tunnelList := strings.FieldsFunc(getEnv("STREMTHRU_TUNNEL", ""), func(c rune) bool {
		return c == ','
	})

	for _, tunnel := range tunnelList {
		if hostname, proxy, ok := strings.Cut(tunnel, ":"); ok {
			if hostname == "*" {
				if proxy == "false" {
					if err := os.Setenv("NO_PROXY", "*"); err != nil {
						log.Fatal("failed to set no_proxy")
					}
				} else if proxy == "true" {
					if err := os.Unsetenv("NO_PROXY"); err != nil {
						log.Fatal("failed to unset no_proxy")
					}
				}
				continue
			}

			switch proxy {
			case "false":
				tunnelMap[hostname] = url.URL{}
			case "true":
				tunnelMap[hostname] = *defaultProxy
			default:
				if u, err := url.Parse(proxy); err == nil {
					tunnelMap[hostname] = *u
				}
			}
		}
	}

	return tunnelMap
}()

// has hostname proxy
var DefaultHTTPTransport = func() *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = Tunnel.GetProxy("hostname")
	transport.DisableKeepAlives = true
	return transport
}()

var DefaultHTTPClient = func() *http.Client {
	transport := DefaultHTTPTransport.Clone()
	return &http.Client{
		Transport: transport,
		Timeout:   90 * time.Second,
	}
}()

func GetHTTPClient(proxyType string) *http.Client {
	transport := DefaultHTTPTransport.Clone()
	transport.Proxy = Tunnel.GetProxy(proxyType)
	return &http.Client{
		Transport: transport,
		Timeout:   90 * time.Second,
	}
}

func getIp(client *http.Client) (string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://checkip.amazonaws.com", nil)
	if err != nil {
		return "", err
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(body)), nil
}

type IPResolver struct {
	machineIP string
}

func (ipr *IPResolver) GetMachineIP() string {
	if ipr.machineIP == "" {
		client := GetHTTPClient("none")
		ip, err := getIp(client)
		if err != nil {
			log.Panicf("Failed to detect Machine IP: %v\n", err)
		}
		ipr.machineIP = ip
	}
	return ipr.machineIP
}

func (ipr *IPResolver) GetTunnelIP() (string, error) {
	client := GetHTTPClient("forced")
	ip, err := getIp(client)
	if err != nil {
		return "", err
	}
	return ip, nil
}
