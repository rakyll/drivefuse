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

package cmd

import (
	"auth"
	"config"
	"fmt"

	"third_party/code.google.com/p/goauth2/oauth"
	"third_party/code.google.com/p/google-api-go-client/drive/v2"

	"logger"
)

const (
    messageWelcome = "Welcome to drived setup and auth wizard."
    messageAddAccount = "Add an account."
)

var ClientIdQuestion = &Question{
	"client_id",
	"OAuth 2.0 Client ID",
	"943748168841.apps.googleusercontent.com",
}

var ClientSecretQuestion = &Question{
	"client_secret",
	"OAuth 2.0 Client Secret",
	"iy1Cbc7CjshE2VqYQ0OfWGxt",
}

var AccountNameQuestion = &Question{
	"account_name",
	"Name for this account.",
	"default_account",
}

var RemoteIdQuestion = &Question{
	"remote_id",
	"Remote folder ID to sync with, L for list.",
	"root",
}

var AuthorizationCodeQuestion = &Question{
	"auth_code",
	"OAuth 2.0 authorization code.",
	"",
}

var LocalPathQuestion = &Question{
	"local_path",
	"Local path to sync from.",
	config.DefaultMountpoint(),
}

func Bold(str string) string {
	return "\033[1m" + str + "\033[0m"
}

func Blue(str string) string {
	return Bold("\033[34m" + str + "\033[0m")
}

type Question struct {
	Name    string
	Usage   string
	Default string
}

func readQuestion(opt *Question) string {
	var s string
	for s == "" {
		fmt.Printf("%v %v [default=%v]>> ", opt.Usage, Bold(opt.Name), Blue(opt.Default))
		_, err := fmt.Scanln(&s)
		if err != nil && err.Error() != "unexpected newline" {
			logger.F("Bad scan.", err)
		}
		if s == "" {
			s = opt.Default
		}
	}
	return s
}

func listFolders(tr *oauth.Transport) {
	svc, err := drive.New(tr.Client())
	if err != nil {
		logger.F(err)
	}
	q := "mimeType='application/vnd.google-apps.folder' and trashed=false"
  // TODO: pagination.
	files, err := svc.Files.List().Q(q).Do()
	if err != nil {
		logger.F(err)
	}
	for _, f := range files.Items {
		fmt.Println(f.Title, f.Id)
	}
}

func retrieveRefreshToken(act *config.Account) string {
	tr := auth.NewTransport(act)
	url := tr.Config.AuthCodeURL("")
	fmt.Println("Visit this URL to get an authorization code.")
	fmt.Println(url)
	code := readQuestion(AuthorizationCodeQuestion)
	token, err := tr.Exchange(code)
	if err != nil {
		logger.F("Failed to exchange authorization code.", err)
	}
	return token.RefreshToken
}

func readAccount() *config.Account {
	cfg := &config.Account{
		LocalPath:    readQuestion(LocalPathQuestion),
		ClientId:     readQuestion(ClientIdQuestion),
		ClientSecret: readQuestion(ClientSecretQuestion),
	}
	cfg.RefreshToken = retrieveRefreshToken(cfg)
	for cfg.RemoteId == "" {
		rid := readQuestion(RemoteIdQuestion)
		if rid == "L" {
			listFolders(auth.NewTransport(cfg))
		} else {
			cfg.RemoteId = rid
		}
	}
	return cfg
}

func readConfig(dataDir string) *config.Config {
	return &config.Config{DataDir: dataDir, Accounts: []*config.Account{readAccount()}}
}

func RunAuthWizard(dataDir string) {
	fmt.Println(messageWelcome)
	fmt.Println(messageAddAccount)
	cfg := readConfig(dataDir)
	err := cfg.Save()
	if err != nil {
		logger.F(err)
	}
	j, err := cfg.Json()
	if err != nil {
		logger.F(err)
	}
	fmt.Println(string(j))
	fmt.Println("Config written to", cfg.ConfigPath())
}
