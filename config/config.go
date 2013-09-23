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
	"os"
  "os/user"
  "path/filepath"

	"github.com/rakyll/drivefuse/logger"
)

const (
	configName = "config.json"
	metaName   = "meta.sql"
	blobName   = "blob"
)

func DefaultMountpoint() string {
	return HomeDir("google-drive")
}

func DefaultDataDir() string {
	return HomeDir(".drived")
}

func HomeDir(path ...string) string {
	usr, err := user.Current()
	if err != nil {
		return ""
	}
	path = append([]string{usr.HomeDir}, path...)
	return filepath.Join(path...)
}

type Account struct {
	LocalPath    string `json:"local_path"`
	RemoteId     string `json:"remote_id"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
}

type Config struct {
  DataDir string `json:"-"` // omits from json marshal/unmarshal
	Accounts []*Account `json:"accounts"`
}

func NewConfig(dataDir string) *Config {
  if dataDir == "" {
    dataDir = DefaultDataDir()
  }
  logger.D("Data directory is", dataDir)
  return &Config{DataDir: dataDir}
}

// Setup initial config requirements - only need some directories for now.
func (c *Config) Setup() error {
  return os.MkdirAll(c.BlobPath(), 0750)
}

func (c *Config) Save() error {
	f, err := os.Create(c.ConfigPath())
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

func (c *Config) Load() error {
	f, err := os.Open(c.ConfigPath())
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
func (c *Config) FirstAccount() *Account {
	return c.Accounts[0]
}

func (c *Config) DataPath(path ...string) string {
	path = append([]string{c.DataDir}, path...)
	return filepath.Join(path...)
}

func (c *Config) ConfigPath() string {
	return c.DataPath(configName)
}

func (c *Config) BlobPath() string {
	return c.DataPath(blobName)
}

func (c *Config) MetadataPath() string {
	return c.DataPath(metaName)
}
