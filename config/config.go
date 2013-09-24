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

// Package config provides configuration management, loading/saving and
// handling.
package config

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/user"
	"path/filepath"

	"github.com/rakyll/drivefuse/logger"
)

const (
	// Name of the configuration file.
	configName = "config.json"

	// Name of the metadata database file.
	metaName = "meta.sql"

	// Name of the blob directory.
	blobName = "blob"
)

// DefaultMountpoint gets the default local path to mount to for a user.
func DefaultMountpoint() string {
	return HomeDir("google-drive")
}

// DefaultDataDir gets the default data directory for a user.
func DefaultDataDir() string {
	return HomeDir(".drived")
}

// HomeDir generates a joined path relative to the user's home directory.
func HomeDir(path ...string) string {
	usr, err := user.Current()
	if err != nil {
		return ""
	}
	path = append([]string{usr.HomeDir}, path...)
	return filepath.Join(path...)
}

// Account is the configuration of a single account.
type Account struct {

	// Local path where a Drive directory will be mounted.
	LocalPath string `json:"local_path"`

	// File ID of the remote folder to be synced.
	RemoteId string `json:"remote_id"`

	// OAuth 2.0 Client ID for authorization and token refreshing.
	ClientId string `json:"client_id"`

	// OAuth 2.0 Client ID for authorization and token refreshing.
	ClientSecret string `json:"client_secret"`

	// OAuth 2.0 refresh token.
	RefreshToken string `json:"refresh_token"`
}

// Validate tests whether all required fields are present.
func (a *Account) Validate() bool {
	return a.LocalPath != "" &&
		a.RemoteId != "" &&
		a.ClientId != "" &&
		a.ClientSecret != "" &&
		a.RefreshToken != ""
}

// Config contains the configuration for the running app.
type Config struct {

	// Base data directory
	DataDir string `json:"-"` // Omits from json marshal/unmarshal.

	// Accounts are the configured accounts.
	Accounts []*Account `json:"accounts"`
}

// NewConfig creates a new configuration in a given directory.
func NewConfig(dataDir string) *Config {
	if dataDir == "" {
		dataDir = DefaultDataDir()
	}
	logger.D("Data directory is", dataDir)
	return &Config{DataDir: dataDir}
}

// Validate tests there is at least one account and all accounts are valid.
func (c *Config) Validate() bool {
	if len(c.Accounts) == 0 {
		return false
	}
	for _, a := range c.Accounts {
		if !a.Validate() {
			return false
		}
	}
	return true
}

// Setup initial config requirements - only need some directories for now.
func (c *Config) Setup() error {
	return os.MkdirAll(c.BlobPath(), 0750)
}

// Save a config file to the configured location in the data directory.
func (c *Config) Save() error {
	f, err := os.Create(c.ConfigPath())
	if err != nil {
		return err
	}
	defer f.Close()
	return c.Write(f)
}

// Marshal the configuration to JSON.
func (c *Config) Marshal() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

// Write the configuration as JSON.
func (c *Config) Write(w io.Writer) error {
	bs, err := c.Marshal()
	if err != nil {
		return err
	}
	_, err = w.Write(bs)
	if err != nil {
		return err
	}
	return nil
}

// Load the configuration from the configured location in the data directory.
func (c *Config) Load() error {
	f, err := os.Open(c.ConfigPath())
	if err != nil {
		return err
	}
	defer f.Close()
	err = c.Read(f)
	if err != nil {
		return err
	}
	valid := c.Validate()
	if !valid {
		return errors.New("Invalid config.")
	}
	return nil
}

// Read the configuration from JSON.
func (c *Config) Read(r io.Reader) error {
	return json.NewDecoder(r).Decode(c)
}

// Hack while we only support one account
func (c *Config) FirstAccount() *Account {
	return c.Accounts[0]
}

// DataPath generates a path relative to the data directory.
func (c *Config) DataPath(path ...string) string {
	path = append([]string{c.DataDir}, path...)
	return filepath.Join(path...)
}

// ConfigPath is the path to the config file inside the data directory.
func (c *Config) ConfigPath() string {
	return c.DataPath(configName)
}

// BlobPath is the path to the blob directory in the data directory.
func (c *Config) BlobPath() string {
	return c.DataPath(blobName)
}

// Metadata path is the path to the metadata database in the data directory.
func (c *Config) MetadataPath() string {
	return c.DataPath(metaName)
}
