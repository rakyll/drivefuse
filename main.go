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
	"log"

	"blob"
	"config"
	"fileio"
	"metadata"
	"mount"
	"syncer"
	client "third_party/code.google.com/p/google-api-go-client/drive/v2"
)

var (
	flagDataPath   = flag.String("datapath", "", "path of the data directory")
	flagMountPoint = flag.String("mountpoint", "", "mount point")
	flagBlockSync  = flag.Bool("blocksync", false, "set true to force blocking sync on startup")

	metaService  *metadata.MetaService
	driveService *client.Service
	blobManager  *blob.Manager
)

func main() {
	flag.Parse()

	cfg, err := config.New(*flagDataPath)
	if err != nil {
		log.Fatal("Error while reading and initializing configuration:", err)
	}

	transport := cfg.GetDefaultTransport()

	metaService, _ = metadata.New(cfg.GetMetadataPath())
	driveService, _ = client.New(transport.Client())
	blobManager = blob.New(cfg.GetBlobPath())

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

	log.Println("mounting...")
	mount.MountAndServe(*flagMountPoint, metaService, blobManager, downloader)
}
