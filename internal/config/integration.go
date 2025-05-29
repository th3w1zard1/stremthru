package config

import (
	"log"
	"net/url"
)

var BaseURL = func() *url.URL {
	baseUrl, err := url.Parse(getEnv("STREMTHRU_BASE_URL"))
	if err != nil {
		log.Panicf("Invalid Base URL: %v\n", err)
	}
	return baseUrl
}()

type integrationConfigTrakt struct {
	ClientId     string
	ClientSecret string
}

func (c integrationConfigTrakt) IsEnabled() bool {
	return c.ClientId != "" && c.ClientSecret != ""
}

type IntegrationConfig struct {
	Trakt integrationConfigTrakt
}

func parseIntegration() IntegrationConfig {
	integration := IntegrationConfig{
		Trakt: integrationConfigTrakt{
			ClientId:     getEnv("STREMTHRU_INTEGRATION_TRAKT_CLIENT_ID"),
			ClientSecret: getEnv("STREMTHRU_INTEGRATION_TRAKT_CLIENT_SECRET"),
		},
	}
	return integration
}

var Integration = parseIntegration()
