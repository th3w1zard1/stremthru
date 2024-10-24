package main

import (
	"encoding/base64"
	"log"
	"os"
	"strings"
)

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

type Config struct {
	Port                string
	EnforceProxyAuth    bool
	ProxyAuthCredential map[string]bool
}

var config = func() Config {
	if value := getEnv("STREMTHRU_HTTP_PROXY", ""); len(value) > 0 {
		if err := os.Setenv("HTTP_PROXY", value); err != nil {
			log.Fatal("failed to set http proxy")
		}
	}

	if value := getEnv("STREMTHRU_HTTPS_PROXY", ""); len(value) > 0 {
		if err := os.Setenv("HTTPS_PROXY", value); err != nil {
			log.Fatal("failed to set https proxy")
		}
	}

	proxyAuthCredList := strings.FieldsFunc(getEnv("STREMTHRU_PROXY_AUTH_CREDENTIALS", ""), func(c rune) bool {
		return c == ','
	})
	proxyAuthCredMap := make(map[string]bool, len(proxyAuthCredList))
	for _, cred := range proxyAuthCredList {
		proxyAuthCredMap[cred] = true
		if strings.ContainsRune(cred, ':') {
			proxyAuthCredMap[base64.StdEncoding.EncodeToString([]byte(cred))] = true
		} else {
			decodedBytes, err := base64.StdEncoding.DecodeString(cred)
			if err == nil {
				proxyAuthCredMap[string(decodedBytes)] = true
			}
		}
	}

	return Config{
		Port:                getEnv("STREMTHRU_PORT", "8080"),
		EnforceProxyAuth:    len(proxyAuthCredList) > 0,
		ProxyAuthCredential: proxyAuthCredMap,
	}
}()
