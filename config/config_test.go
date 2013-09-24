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
	"os"
	"path/filepath"
	"testing"

	T "github.com/rakyll/drivefuse/third_party/launchpad.net/gocheck"
)

// Create the test suite
type ConfigSuite struct {
	dataDir string
}

func (s *ConfigSuite) SetUpTest(c *T.C) {
	s.dataDir = c.MkDir()
}

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) {
	T.Suite(&ConfigSuite{})
	T.TestingT(t)
}

type fileExistsChecker struct {
	*T.CheckerInfo
}

func (checker *fileExistsChecker) Check(params []interface{}, names []string) (bool, string) {
	_, err := os.Stat(params[0].(string))
	if err != nil {
		if os.IsNotExist(err) {
			return false, "File does not exist."
		} else {
			return false, err.Error()
		}
	}
	return true, ""
}

var fileExists T.Checker = &fileExistsChecker{
	&T.CheckerInfo{Name: "FileExists", Params: []string{"path"}},
}

var testGoodFile string = `
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

var testMissingAccountAttribute string = `
{
  "accounts": [
    {
      "remote_id": "root",
      "client_id": "943748168841.apps.googleusercontent.com",
      "client_secret": "iy1Cbc7CjshE2VqYQ0OfWGxt",
      "refresh_token": "1/Hm2qp_5zZxhMH8mIo1-XGE24f_XtL3-PdV749nHzz6Q"
    }
  ]
}
`

var testNoAccountsFile string = `
{
  "accounts": []
}
`

func (s *ConfigSuite) TestNewConfig(c *T.C) {
	cfg := NewConfig(s.dataDir)
	c.Assert(s.dataDir, T.Equals, cfg.DataDir)
}

func (s *ConfigSuite) TestConfigSetup(c *T.C) {
	cfg := NewConfig(s.dataDir)
	cfg.Setup()
	c.Assert(filepath.Join(s.dataDir, blobName), fileExists)
}

func (s *ConfigSuite) TestConfigPath(c *T.C) {
	cfg := NewConfig(s.dataDir)
	c.Assert(filepath.Join(s.dataDir, configName), T.Equals, cfg.ConfigPath())
}

func (s *ConfigSuite) TestGoodConfigLoad(c *T.C) {
	cfg := NewConfig(s.dataDir)
	cfg.Setup()
	f, err := os.Create(filepath.Join(s.dataDir, configName))
	if err != nil {
		c.Error(err)
	}
	f.WriteString(testGoodFile)
	cfg.Load()
	c.Assert("iy1Cbc7CjshE2VqYQ0OfWGxt", T.Equals, cfg.FirstAccount().ClientSecret)
	// Let's just say json unmarshalling works
}

func (s *ConfigSuite) TestMissingAccountField(c *T.C) {
	cfg := NewConfig(s.dataDir)
	cfg.Setup()
	f, err := os.Create(filepath.Join(s.dataDir, configName))
	if err != nil {
		c.Error(err)
	}
	f.WriteString(testMissingAccountAttribute)
	err = cfg.Load()
	c.Assert(err, T.NotNil)
}

func (s *ConfigSuite) TestZeroAccounts(c *T.C) {
	cfg := NewConfig(s.dataDir)
	cfg.Setup()
	f, err := os.Create(filepath.Join(s.dataDir, configName))
	if err != nil {
		c.Error(err)
	}
	f.WriteString(testNoAccountsFile)
	err = cfg.Load()
	c.Assert(err, T.NotNil)
}

func (s *ConfigSuite) TestDataDirPath(c *T.C) {
	cfg := NewConfig(s.dataDir)
	c.Assert(filepath.Join(s.dataDir, configName), T.Equals, cfg.ConfigPath())
}

func (s *ConfigSuite) TestFailing(c *T.C) {
	c.Error(1)
}
