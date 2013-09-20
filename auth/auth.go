package auth

import (
	"config"
	"net/http"
	"third_party/code.google.com/p/goauth2/oauth"
	"time"
)

const (
	GoogleOAuth2AuthURL  string = "https://accounts.google.com/o/oauth2/auth"
	GoogleOAuth2TokenURL string = "https://accounts.google.com/o/oauth2/token"
	RedirectURL          string = "urn:ietf:wg:oauth:2.0:oob"
	DriveScope           string = "https://www.googleapis.com/auth/drive"
	AccessType           string = "offline"
)

func ClientConfig(cfg *config.AccountConfig) *oauth.Config {
	return &oauth.Config{
		ClientId:     cfg.ClientId,
		ClientSecret: cfg.ClientSecret,
		AuthURL:      GoogleOAuth2AuthURL,
		TokenURL:     GoogleOAuth2TokenURL,
	}
}

func AuthConfig(cfg *config.AccountConfig) *oauth.Config {
	return &oauth.Config{
		ClientId:     cfg.ClientId,
		ClientSecret: cfg.ClientSecret,
		AuthURL:      GoogleOAuth2AuthURL,
		TokenURL:     GoogleOAuth2TokenURL,
		RedirectURL:  RedirectURL,
		AccessType:   AccessType,
		Scope:        DriveScope,
	}
}

func Transport(cfg *config.AccountConfig, oauthConfig *oauth.Config) *oauth.Transport {
	// force refreshes the access token on start, make sure
	// refresh request in parallel are being started
	return &oauth.Transport{
		Config:    oauthConfig,
		Transport: http.DefaultTransport,
		Token:     &oauth.Token{RefreshToken: cfg.RefreshToken, Expiry: time.Now()},
	}
}

func ClientTransport(cfg *config.AccountConfig) *oauth.Transport {
	return Transport(cfg, ClientConfig(cfg))
}
