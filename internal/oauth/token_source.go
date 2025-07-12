package oauth

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
)

type TokenSourceConfig struct {
	Provider     Provider
	GetUser      func(client *http.Client, oauthConfig *oauth2.Config) (userId, userName string, err error)
	PrepareToken func(tok *oauth2.Token, id, userId, userName string) *oauth2.Token
}

type dbTokenSource struct {
	oauth2.TokenSource
	oauthConfig *oauth2.Config
	config      TokenSourceConfig
	TokenId     string
	tok         *oauth2.Token
}

func (ts *dbTokenSource) save() error {
	otok := &OAuthToken{}
	otok = otok.FromToken(ts.tok)
	tokenSourceLog.Debug("saving token", "provider", ts.config.Provider, "user_id", otok.UserId, "user_name", otok.UserName)
	return SaveOAuthToken(otok)
}

func (ts *dbTokenSource) Token() (*oauth2.Token, error) {
	if ts.tok.Valid() {
		return ts.tok, nil
	}

	tokenSourceLog.Debug("token expired, refreshing token", "provider", ts.config.Provider, "user_id", ts.tok.Extra("user_id"), "user_name", ts.tok.Extra("user_name"))
	tok, err := ts.TokenSource.Token()
	if err != nil {
		tokenSourceLog.Error("failed to refresh token", "error", err, "provider", ts.config.Provider, "user_id", ts.tok.Extra("user_id"), "user_name", ts.tok.Extra("user_name"))
		return nil, err
	}

	userId, userName, err := ts.config.GetUser(
		oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(tok)),
		ts.oauthConfig,
	)
	if err != nil {
		return nil, err
	}
	ts.tok = ts.config.PrepareToken(tok, ts.TokenId, userId, userName)
	err = ts.save()
	if err != nil {
		return nil, err
	}

	return ts.tok, nil
}

type DatabaseTokenSourceConfig struct {
	OAuth *oauth2.Config
	TokenSourceConfig
}

func DatabaseTokenSource(conf *DatabaseTokenSourceConfig, token *oauth2.Token) oauth2.TokenSource {
	return oauth2.ReuseTokenSource(token, &dbTokenSource{
		TokenSource: conf.OAuth.TokenSource(context.Background(), token),
		config:      conf.TokenSourceConfig,
		oauthConfig: conf.OAuth,
		TokenId:     token.Extra("id").(string),
		tok:         token,
	})
}
