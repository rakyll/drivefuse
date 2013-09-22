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

package fileio

import (
	"net/http"
	"sync"
	"time"

	"github.com/rakyll/drivefuse/blob"
	"github.com/rakyll/drivefuse/logger"
	"github.com/rakyll/drivefuse/metadata"
)

const (
	IntervalDownloadTicker         = 5 * time.Second // TODO(burcud): need to be adaptive
	MaxNumberOfConcurrentDownloads = 10

	BaseUrlDownloadHost = "https://googledrive.com/host"
)

type Downloader struct {
	client      *http.Client
	metaService *metadata.MetaService
	blobMngr    *blob.Manager

	mu sync.Mutex
}

func NewDownloader(client *http.Client, m *metadata.MetaService, blobMngr *blob.Manager) *Downloader {
	downloader := &Downloader{
		client:      client,
		metaService: m,
		blobMngr:    blobMngr,
	}
	downloader.Start()
	return downloader
}

func (d *Downloader) Start() {
	go func() {
		for {
			d.tick()
			<-time.After(IntervalDownloadTicker)
		}
	}()
}

func (d *Downloader) tick() {
	d.mu.Lock()
	defer d.mu.Unlock()
	// TODO: add an additional queue for small sized files
	// so that, large files dont block the download queue.
	// retrieve at least MaxNumberOfConcurrentDownloads files to download
	downloads, _ := d.metaService.ListDownloads(MaxNumberOfConcurrentDownloads)
	if len(downloads) == 0 {
		return
	}
	completed := make(chan bool, len(downloads))
	for _, item := range downloads {
		go func(id string, checksum string, ch chan bool) {
			d.download(id, checksum)
			ch <- true
		}(item.Id, item.Md5Checksum, completed)
	}
	<-completed
}

func (d *Downloader) download(id string, checksum string) {
	// TODO: handle all error cases, make sure queue is not blocked
	// with erroneous files
	logger.V("Downloading", id, checksum)
	var (
		resp *http.Response
		err  error
	)
	if resp, err = d.client.Get(BaseUrlDownloadHost + "/" + id); err != nil {
		logger.V("error downloading", id, err)
		return
	}

	if resp.StatusCode == 404 {
		d.metaService.DequeueFromIO("download", id)
		logger.V("error downloading [not found]", id)
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		logger.V("error downloading [not ok]", id, resp.StatusCode)
		return
	}

	defer resp.Body.Close()
	err = d.blobMngr.Save(id, checksum, resp.Body)
	if err != nil {
		logger.V(err)
		return
	}

	err = d.metaService.InitFile(id)
	if err != nil {
		logger.V(err)
		return
	}

	d.metaService.DequeueFromIO("download", id)
}
