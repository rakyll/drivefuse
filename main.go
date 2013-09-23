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

package main

import (
	"flag"
	"os"

	"github.com/rakyll/drivefuse/auth"
	"github.com/rakyll/drivefuse/cmd"
	"github.com/rakyll/drivefuse/blob"
	"github.com/rakyll/drivefuse/config"
	"github.com/rakyll/drivefuse/fileio"
	"github.com/rakyll/drivefuse/logger"
	"github.com/rakyll/drivefuse/metadata"
	"github.com/rakyll/drivefuse/mount"
	"github.com/rakyll/drivefuse/syncer"

	client "github.com/rakyll/drivefuse/third_party/code.google.com/p/google-api-go-client/drive/v2"
)

var (
	flagDataDir   = flag.String("datadir", config.DefaultDataDir(), "path of the data directory")
	flagMountPoint = flag.String("mountpoint", config.DefaultMountpoint(), "mount point")
	flagBlockSync  = flag.Bool("blocksync", false, "set true to force blocking sync on startup")

	flagRunAuthWizard = flag.Bool("wizard", false, "Run the startup wizard.")

	metaService  *metadata.MetaService
	driveService *client.Service
	blobManager  *blob.Manager
)

func main() {
	flag.Parse()

	if *flagRunAuthWizard {
		cmd.RunAuthWizard(*flagDataDir)
		os.Exit(0)
	}

	cfg := config.NewConfig(*flagDataDir)
  err := cfg.Setup()
	if err != nil {
		logger.F("Error initializing configuration.", err)
	}
	err = cfg.Load()
	if err != nil {
		logger.F("Did you mean --wizard? Error reading configuration.", err)
	}

	transport := auth.NewTransport(cfg.FirstAccount())

	metaService, _ = metadata.New(cfg.MetadataPath())
	driveService, _ = client.New(transport.Client())
	blobManager = blob.New(cfg.BlobPath())

	downloader := fileio.NewDownloader(
		transport.Client(),
		metaService,
		blobManager)

	syncManager := syncer.NewCachedSyncer(
		driveService,
		metaService,
		blobManager)

	if *flagBlockSync {
		syncManager.Sync(true)
	}
	syncManager.Start()

	logger.V("mounting...")
	mountpoint := cfg.FirstAccount().LocalPath
	err = os.MkdirAll(mountpoint, 0774)
	if err != nil {
		logger.F(err)
	}
	if err = mount.MountAndServe(mountpoint, metaService, blobManager, downloader); err != nil {
		logger.F(err)
	}
}
