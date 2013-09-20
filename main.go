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
	"auth"
	"blob"
	"config"
	"fileio"
	"flag"
	"logger"
	"metadata"
	"mount"
	"os"
	"syncer"
	client "third_party/code.google.com/p/google-api-go-client/drive/v2"
	"wizard"
)

var (
	flagDataPath   = flag.String("datapath", config.DefaultDatadir(), "path of the data directory")
	flagMountPoint = flag.String("mountpoint", "", "mount point")
	flagBlockSync  = flag.Bool("blocksync", false, "set true to force blocking sync on startup")

	flagWizard = flag.Bool("wizard", false, "Run the startup wizard.")

	metaService  *metadata.MetaService
	driveService *client.Service
	blobManager  *blob.Manager
)

func main() {
	flag.Parse()

	env, err := config.NewEnv(*flagDataPath)
	if *flagWizard {
		wizard.Run(env)
		os.Exit(0)
	}

	err = env.LoadConfig()
	if err != nil {
		logger.F("Error while reading and initializing configuration:", err)
	}

	transport := auth.ClientTransport(env.Config.FirstAccount())

	metaService, _ = metadata.New(env.MetadataPath())
	driveService, _ = client.New(transport.Client())
	blobManager = blob.New(env.BlobPath())

	downloader := fileio.NewDownloader(
		transport.Client(),
		metaService,
		blobManager)

	syncManager := syncer.NewCachedSyncer(
		driveService,
		metaService)

	if *flagBlockSync {
		syncManager.Sync(true)
	}
	syncManager.Start()

	logger.V("mounting...")
	mountpoint := env.Config.FirstAccount().LocalPath
	err = os.MkdirAll(mountpoint, 0774)
	if err != nil {
		logger.F(err)
	}
	if err = mount.MountAndServe(mountpoint, metaService, blobManager, downloader); err != nil {
		logger.F(err)
	}
}
