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

package wizard

import (
	"auth"
	"config"
	"fmt"
	"log"
	"third_party/code.google.com/p/goauth2/oauth"
	"third_party/code.google.com/p/google-api-go-client/drive/v2"
)

const motd1 string = "Welcome to drived setup and auth wizard."
const motd2 string = "Add an account."

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

func ReadQuestion(opt *Question) string {
	var s string
	for s == "" {
		fmt.Printf("%v %v [default=%v]>> ", opt.Usage, Bold(opt.Name), Blue(opt.Default))
		_, err := fmt.Scanln(&s)
		if err != nil && err.Error() != "unexpected newline" {
			log.Fatalln("Bad scan", err)
		}
		if s == "" {
			s = opt.Default
		}
	}
	return s
}

func ListFolders(tr *oauth.Transport) {
	svc, err := drive.New(tr.Client())
	if err != nil {
		log.Fatalln(err)
	}
	q := "mimeType='application/vnd.google-apps.folder' and trashed=false"
	files, err := svc.Files.List().Q(q).Do()
	if err != nil {
		log.Fatalln(err)
	}
	for _, f := range files.Items {
		fmt.Println(f.Title, f.Id)
	}
}

func Auth(cfg *config.AccountConfig) string {
	c := auth.AuthConfig(cfg)
	tr := &oauth.Transport{Config: c}
	url := c.AuthCodeURL("")
	fmt.Println("Visit this URL to get a code.")
	fmt.Println(url)
	code := ReadQuestion(AuthorizationCodeQuestion)
	token, err := tr.Exchange(code)
	if err != nil {
		log.Fatalln("Failed to exchange authorization code.", err)
	}
	return token.RefreshToken
}

func ReadAccount() *config.AccountConfig {
	cfg := &config.AccountConfig{
		AccountName:  ReadQuestion(AccountNameQuestion),
		LocalPath:    ReadQuestion(LocalPathQuestion),
		ClientId:     ReadQuestion(ClientIdQuestion),
		ClientSecret: ReadQuestion(ClientSecretQuestion),
	}
	cfg.RefreshToken = Auth(cfg)
	for cfg.RemoteId == "" {
		rid := ReadQuestion(RemoteIdQuestion)
		if rid == "L" {
			ListFolders(auth.ClientTransport(cfg))
		} else {
			cfg.RemoteId = rid
		}
	}
	return cfg
}

func ReadConfig() *config.Config {
	return &config.Config{Accounts: []*config.AccountConfig{ReadAccount()}}
}

func Run(env *config.Env) {
	log.Println(motd1)
	log.Println(motd2)
	cfg := ReadConfig()
	err := cfg.Save(env.ConfigPath())
	if err != nil {
		log.Fatalln(err)
	}
	j, err := cfg.Json()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(string(j))
	log.Println("Config written to", env.ConfigPath())
}
