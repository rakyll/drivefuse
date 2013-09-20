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

package config

import (
	"encoding/json"
	"io"
	_ "logger"
	"os"
)

type AccountConfig struct {
	AccountName  string `json:"account_name"`
	LocalPath    string `json:"local_path"`
	RemoteId     string `json:"remote_id"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
}

type Config struct {
	Accounts []*AccountConfig `json:"accounts"`
}

func (c *Config) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return c.Write(f)
}

func (c *Config) Json() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

func (c *Config) Write(w io.Writer) error {
	bs, err := c.Json()
	if err != nil {
		return err
	}
	_, err = w.Write(bs)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) Load(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return c.Read(f)
}

func (c *Config) Read(r io.Reader) error {
	return json.NewDecoder(r).Decode(c)
}

// Hack while we only support one account
func (c *Config) FirstAccount() *AccountConfig {
	return c.Accounts[0]
}
