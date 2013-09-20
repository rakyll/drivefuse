package config

import (
	"os"
	"os/user"
	"path/filepath"
)

const (
	configName = "config.json"
	metaName   = "meta.sql"
	blobName   = "blob"
)

func DefaultMountpoint() string {
	return HomeDir("google-drive")
}

func DefaultDatadir() string {
	return HomeDir(".drived")
}

type Env struct {
	DataDir string
	Config  *Config
}

func NewEnv(dataDir string) (*Env, error) {
	env := &Env{DataDir: dataDir, Config: &Config{}}
	if err := os.MkdirAll(env.BlobPath(), 0750); err != nil {
		return nil, err
	}
	return env, nil
}

func (e *Env) LoadConfig() error {
	return e.Config.Load(e.ConfigPath())
}

func (e *Env) DataPath(path ...string) string {
	path = append([]string{e.DataDir}, path...)
	return filepath.Join(path...)
}

func (e *Env) ConfigPath() string {
	return e.DataPath(configName)
}

func (e *Env) BlobPath() string {
	return e.DataPath(blobName)
}

func (e *Env) MetadataPath() string {
	return e.DataPath(metaName)
}

func HomeDir(path ...string) string {
	usr, err := user.Current()
	if err != nil {
		return ""
	}
	path = append([]string{usr.HomeDir}, path...)
	return filepath.Join(path...)
}
