package config

import (
	"log"
	"net/url"
	"strings"
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

type integrationConfigKitsu struct {
	ClientId     string
	ClientSecret string
	Email        string
	Password     string
}

func (c integrationConfigKitsu) HasDefaultCredentials() bool {
	return c.Email != "" && c.Password != ""
}

type IntegrationConfig struct {
	Trakt integrationConfigTrakt
	Kitsu integrationConfigKitsu
}

func parseIntegration() IntegrationConfig {
	integration := IntegrationConfig{
		Trakt: integrationConfigTrakt{
			ClientId:     getEnv("STREMTHRU_INTEGRATION_TRAKT_CLIENT_ID"),
			ClientSecret: getEnv("STREMTHRU_INTEGRATION_TRAKT_CLIENT_SECRET"),
		},
		Kitsu: integrationConfigKitsu{
			ClientId:     getEnv("STREMTHRU_INTEGRATION_KITSU_CLIENT_ID"),
			ClientSecret: getEnv("STREMTHRU_INTEGRATION_KITSU_CLIENT_SECRET"),
			Email:        getEnv("STREMTHRU_INTEGRATION_KITSU_EMAIL"),
			Password:     getEnv("STREMTHRU_INTEGRATION_KITSU_PASSWORD"),
		},
	}
	if integration.Kitsu.Email != "" && !strings.Contains(integration.Kitsu.Email, "@") {
		log.Panicf("Invalid Kitsu Email: %s\n", integration.Kitsu.Email)
	}
	return integration
}

var Integration = parseIntegration()
