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

	downloader := fileio.NewDownloader(transport.Client(), metaService, blobManager)
	syncManager := syncer.New(driveService, metaService)

	if *flagBlockSync {
		syncManager.Sync(true)
	}
	syncManager.Start()

	log.Println("mounting...")
	mount.MountAndServe(*flagMountPoint, metaService, blobManager, downloader)
}
