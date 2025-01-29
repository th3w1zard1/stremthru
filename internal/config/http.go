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

type TunnelType string

const (
	TUNNEL_TYPE_NONE   TunnelType = ""
	TUNNEL_TYPE_AUTO   TunnelType = "a"
	TUNNEL_TYPE_FORCED TunnelType = "f"
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
func (tm TunnelMap) autoProxy(r *http.Request) (*url.URL, error) {
	proxy := tm.getProxy(r.URL.Hostname())
	if proxy == nil {
		return http.ProxyFromEnvironment(r)
	}
	if proxy.Host == "" {
		return nil, nil
	}
	return proxy, nil
}

// Use the default tunnel, ignore `NO_PROXY`
func (tm TunnelMap) forcedProxy(r *http.Request) (*url.URL, error) {
	if proxy := tm.getProxy("*"); proxy != nil && proxy.Host != "" {
		return proxy, nil
	}
	return nil, nil
}

func (tm TunnelMap) GetProxy(tunnelType TunnelType) func(req *http.Request) (*url.URL, error) {
	switch tunnelType {
	case TUNNEL_TYPE_AUTO:
		return tm.autoProxy
	case TUNNEL_TYPE_FORCED:
		return tm.forcedProxy
	case TUNNEL_TYPE_NONE:
		return nil
	default:
		panic("invalid tunnel type")
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

type StoreTunnelConfig struct {
	api    bool
	stream bool
}

type StoreTunnelConfigMap map[string]StoreTunnelConfig

func (stc StoreTunnelConfigMap) isEnabledForAPI(name string) bool {
	if c, ok := stc[name]; ok {
		return c.api
	}
	if name != "*" {
		return stc.isEnabledForAPI("*")
	}
	return true
}

func (stc StoreTunnelConfigMap) GetTypeForAPI(name string) TunnelType {
	enabled := stc.isEnabledForAPI(name)
	if enabled {
		return TUNNEL_TYPE_FORCED
	}
	return TUNNEL_TYPE_NONE
}

func (stc StoreTunnelConfigMap) isEnabledForStream(name string) bool {
	if c, ok := stc[name]; ok {
		return c.stream
	}
	if name != "*" {
		return stc.isEnabledForStream("*")
	}
	return true
}

func (stc StoreTunnelConfigMap) GetTypeForStream(name string) TunnelType {
	enabled := stc.isEnabledForStream(name)
	if enabled {
		return TUNNEL_TYPE_FORCED
	}
	return TUNNEL_TYPE_NONE
}

var StoreTunnel = func() StoreTunnelConfigMap {
	storeTunnelList := strings.FieldsFunc(getEnv("STREMTHRU_STORE_TUNNEL", "*:true"), func(c rune) bool {
		return c == ','
	})

	contentHostnameByStore := map[string]string{
		"alldebrid":  "debrid.it",
		"debridlink": "debrid.link",
		"premiumize": "energycdn.com",
		"realdebrid": "download.real-debrid.com",
		"torbox":     "tb-cdn.st",
	}

	storeTunnelMap := make(StoreTunnelConfigMap)
	for _, storeTunnel := range storeTunnelList {
		if store, tunnel, ok := strings.Cut(storeTunnel, ":"); ok {
			storeTunnelMap[store] = StoreTunnelConfig{
				api:    tunnel == "true" || tunnel == "api",
				stream: tunnel == "true",
			}

			switch store {
			case "*":
				for _, hostname := range contentHostnameByStore {
					if _, exists := Tunnel[hostname]; !exists {
						if tunnel == "true" {
							Tunnel[hostname] = *Tunnel.getProxy("*")
						} else {
							Tunnel[hostname] = url.URL{}
						}
					}
				}
			default:
				if hostname, ok := contentHostnameByStore[store]; ok {
					if tunnel == "true" {
						Tunnel[hostname] = *Tunnel.getProxy("*")
					} else {
						Tunnel[hostname] = url.URL{}
					}
				}
			}
		}
	}

	return storeTunnelMap
}()

// has auto proxy
var DefaultHTTPTransport = func() *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = Tunnel.GetProxy(TUNNEL_TYPE_AUTO)
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

func GetHTTPClient(tunnelType TunnelType) *http.Client {
	transport := DefaultHTTPTransport.Clone()
	transport.Proxy = Tunnel.GetProxy(tunnelType)
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
		client := GetHTTPClient(TUNNEL_TYPE_NONE)
		ip, err := getIp(client)
		if err != nil {
			log.Panicf("Failed to detect Machine IP: %v\n", err)
		}
		ipr.machineIP = ip
	}
	return ipr.machineIP
}

func (ipr *IPResolver) GetTunnelIP() (string, error) {
	client := GetHTTPClient(TUNNEL_TYPE_FORCED)
	ip, err := getIp(client)
	if err != nil {
		return "", err
	}
	return ip, nil
}
