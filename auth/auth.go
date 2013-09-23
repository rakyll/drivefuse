// Copyright 2013 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

import (
	"net/http"
	"time"

	"github.com/rakyll/drivefuse/config"
	"github.com/rakyll/drivefuse/third_party/code.google.com/p/goauth2/oauth"
)

const (
	GoogleOAuth2AuthURL  string = "https://accounts.google.com/o/oauth2/auth"
	GoogleOAuth2TokenURL string = "https://accounts.google.com/o/oauth2/token"
	RedirectURL          string = "urn:ietf:wg:oauth:2.0:oob"
	DriveScope           string = "https://www.googleapis.com/auth/drive"
	AccessType           string = "offline"
)

func newConfig(cfg *config.Account) *oauth.Config {
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

func NewTransport(cfg *config.Account) *oauth.Transport {
	return &oauth.Transport{
		Config:    newConfig(cfg),
		Transport: http.DefaultTransport,
		Token:     &oauth.Token{RefreshToken: cfg.RefreshToken, Expiry: time.Now()},
	}
}
