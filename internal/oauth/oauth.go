package oauth

import "golang.org/x/oauth2"

type OAuthConfig struct {
	oauth2.Config
	Exchange                 func(code, state string) (*oauth2.Token, error)
	PasswordCredentialsToken func(username, password string) (*oauth2.Token, error)
}
