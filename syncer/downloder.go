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

package syncer

import (
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/rakyll/drivefuse/blob"
	"github.com/rakyll/drivefuse/logger"
	"github.com/rakyll/drivefuse/metadata"
)

const (
	intervalTick                           = 5 * time.Second // TODO(burcud): need to be adaptive
	maxNumberOfConcurrentDownloadsPerQueue = 5
	maxSizeQueueTreshold                   = 1 << 20 // TODO(burcud): need to be adaptive

	baseUrlDownloadHost = "https://googledrive.com/host"
)

type Downloader struct {
	client      *http.Client
	metaService *metadata.MetaService
	blobMngr    *blob.Manager

	muSmall sync.Mutex
	muLarge sync.Mutex
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
			d.tickForSmall()
			<-time.After(intervalTick)
		}
	}()
	go func() {
		for {
			d.tickForLarge()
			<-time.After(intervalTick)
		}
	}()
}

func (d *Downloader) tickForSmall() {
	d.muSmall.Lock()
	defer d.muSmall.Unlock()
	d.tick(0, maxSizeQueueTreshold)
}

func (d *Downloader) tickForLarge() {
	d.muLarge.Lock()
	defer d.muLarge.Unlock()
	d.tick(maxSizeQueueTreshold+1, math.MaxInt64)
}

func (d *Downloader) tick(minSize int64, maxSize int64) {
	// TODO: add an additional queue for small sized files
	// so that, large files dont block the download queue.
	// retrieve at least MaxNumberOfConcurrentDownloads files to download
	downloads, _ := d.metaService.ListDownloads(maxNumberOfConcurrentDownloadsPerQueue, minSize, maxSize)
	if len(downloads) == 0 {
		return
	}
	completed := make(chan bool, len(downloads))
	for _, item := range downloads {
		go func(localId int64, remoteId string, checksum string, ch chan bool) {
			d.download(localId, remoteId, checksum)
			ch <- true
		}(item.LocalId, item.Id, item.Md5Checksum, completed)
	}
	<-completed
}

func (d *Downloader) download(localId int64, remoteId string, checksum string) {
	// TODO: handle all error cases, make sure queue is not blocked
	// with erroneous files
	logger.V("Downloading", remoteId, checksum)
	var (
		resp *http.Response
		err  error
	)
	if resp, err = d.client.Get(baseUrlDownloadHost + "/" + remoteId); err != nil {
		logger.V("error downloading", remoteId, err)
		return
	}

	if resp.StatusCode == 404 {
		d.metaService.SetOp(localId, metadata.OpNone)
		logger.V("error downloading [not found]", remoteId)
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		logger.V("error downloading [not ok]", remoteId, resp.StatusCode)
		return
	}

	defer resp.Body.Close()
	err = d.blobMngr.Save(localId, checksum, resp.Body)
	if err != nil {
		logger.V(err)
		return
	}

	d.metaService.SetOp(localId, metadata.OpNone)
}
