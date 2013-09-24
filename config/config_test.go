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

// Contains tests for config package.
package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type configTestVars struct {
	T *testing.T
	DataDir string
}

func newDataDir(t *testing.T) string {
	name, err := ioutil.TempDir("/tmp", "drived-config-text")
	if err != nil {
		t.Error(err)
	}
	return name
}

func setup(t *testing.T) *configTestVars {
	dataDir := newDataDir(t)
	return &configTestVars{
		T: t,
		DataDir: dataDir,
	}
}

func tearDown(v *configTestVars) {
	err := os.RemoveAll(v.DataDir)
	if err != nil {
		v.T.Error(err)
	}
}

func failIfNotExist(t *testing.T, path string) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			t.Errorf("Assert, does not exist (%v) %v", path, err)
		} else {
			t.Error(err)
		}
	}
}

func failIfNotEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Assert equal, expecting (%v) got (%v)", a, b)
	}
}

var testFile string = `
{
  "accounts": [
    {
      "local_path": "/usr/local/google/home/afshar/google-drive",
      "remote_id": "root",
      "client_id": "943748168841.apps.googleusercontent.com",
      "client_secret": "iy1Cbc7CjshE2VqYQ0OfWGxt",
      "refresh_token": "1/Hm2qp_5zZxhMH8mIo1-XGE24f_XtL3-PdV749nHzz6Q"
    }
  ]
}
`

func TestNewConfig(t *testing.T) {
	v := setup(t)
	defer tearDown(v)
	cfg := NewConfig(v.DataDir)
	failIfNotEqual(t, v.DataDir, cfg.DataDir)
}


func TestConfigSetup(t *testing.T) {
	v := setup(t)
	defer tearDown(v)
	cfg := NewConfig(v.DataDir)
	cfg.Setup()
	failIfNotExist(t, filepath.Join(v.DataDir, blobName))
}

func TestConfigPath(t *testing.T) {
	v := setup(t)
	defer tearDown(v)
	cfg := NewConfig(v.DataDir)
	failIfNotEqual(t, filepath.Join(v.DataDir, configName), cfg.ConfigPath())

}

func TestConfigLoad(t *testing.T) {
	v := setup(t)
	defer tearDown(v)
	cfg := NewConfig(v.DataDir)
	cfg.Setup()
	f, err := os.Create(filepath.Join(v.DataDir, configName))
	if err != nil {
		t.Error(err)
	}
	f.WriteString(testFile)
	cfg.Load()
	failIfNotEqual(t, cfg.FirstAccount().ClientSecret, "iy1Cbc7CjshE2VqYQ0OfWGxt")
	// Let's just say json unmarshalling works
}

func TestDataDirPath(t *testing.T) {
	v := setup(t)
	cfg := NewConfig(v.DataDir)
	defer tearDown(v)
	failIfNotEqual(t, filepath.Join(v.DataDir, configName), cfg.ConfigPath())
}

func TestFailing(t *testing.T) {
}
