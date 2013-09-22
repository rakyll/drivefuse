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
	"sync"
	"time"

	"github.com/rakyll/drivefuse/blob"
	"github.com/rakyll/drivefuse/logger"
	"github.com/rakyll/drivefuse/metadata"
	client "github.com/rakyll/drivefuse/third_party/code.google.com/p/google-api-go-client/drive/v2"
)

const (
	intervalSync   = 30 * time.Second // TODO: should be adaptive
	layoutDateTime = "2013-09-19T14:29:12.570Z"
)

type CachedSyncer struct {
	remoteService *client.Service
	metaService   *metadata.MetaService
	blobManager   *blob.Manager

	mu sync.RWMutex
}

func NewCachedSyncer(service *client.Service, metaService *metadata.MetaService, blobManager *blob.Manager) *CachedSyncer {
	return &CachedSyncer{
		remoteService: service,
		metaService:   metaService,
		blobManager:   blobManager,
	}
}

func (d *CachedSyncer) Start() {
	go func() {
		for {
			d.Sync(false)
			<-time.After(intervalSync)
		}
	}()
}

func (d *CachedSyncer) Sync(isForce bool) (err error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	logger.V("Started syncer...")
	err = d.syncInbound(isForce)
	if err != nil {
		logger.V("error during sync", err)
		return
	}
	logger.V("Done syncing...")
	return
}

func (d *CachedSyncer) syncOutbound(rootId string, isRecursive bool, isForce bool) error {
	panic("not implemented")
	return nil
}

func (d *CachedSyncer) syncInbound(isForce bool) (err error) {
	var largestChangeId int64
	largestChangeId, err = d.metaService.GetLargestChangeId()
	isInitialSync := largestChangeId == 0
	if isForce || err != nil {
		largestChangeId = 0
	} else {
		largestChangeId += 1
	}

	// retrieve metadata about root
	var rootFile *client.File
	if rootFile, err = d.remoteService.Files.Get(metadata.IdRootFolder).Do(); err != nil {
		return
	}

	data := buildMetadata(metadata.IdRootFolder, "", rootFile)
	if err = d.metaService.Save("", metadata.IdRootFolder, data, false, false); err != nil {
		return
	}
	pageToken := ""
	for {
		pageToken, err = d.mergeChanges(isInitialSync, rootFile.Id, largestChangeId, pageToken)
		if err != nil || pageToken == "" {
			return
		}
	}
	return
}

func (d *CachedSyncer) mergeChanges(isInitialSync bool, rootId string, startChangeId int64, pageToken string) (nextPageToken string, err error) {
	logger.V("merging changes starting with pageToken:", pageToken, "and startChangeId", startChangeId)

	req := d.remoteService.Changes.List()
	req.IncludeSubscribed(false)
	if pageToken != "" {
		req.PageToken(pageToken)
	} else if startChangeId > 0 { // can't set page token and start change mutually
		req.StartChangeId(startChangeId)
	}
	if isInitialSync {
		req.IncludeDeleted(false)
	}

	var changes *client.ChangeList
	if changes, err = req.Do(); err != nil {
		return
	}

	var largestId int64
	nextPageToken = changes.NextPageToken
	for _, item := range changes.Items {
		if err = d.mergeChange(rootId, item); err != nil {
			return
		}
		largestId = item.Id
	}
	if largestId > 0 {
		// persist largest change id
		d.metaService.SaveLargestChangeId(largestId)
	}
	return
}

func (d *CachedSyncer) mergeChange(rootId string, item *client.Change) (err error) {
	if item.Deleted || item.File.Labels.Trashed {
		// TODO(burcud): Handle directory deletions
		if d.metaService.Delete(item.FileId); err != nil {
			return
		}
		// delete contents
		if d.blobManager.Delete(item.FileId); err != nil {
			return
		}
	} else {
		if item.File.DownloadUrl == "" && item.File.MimeType != metadata.MimeTypeFolder {
			return
		}

		fileId := item.FileId
		parentId := ""
		if len(item.File.Parents) > 0 {
			parentId = item.File.Parents[0].Id
		}
		if parentId == rootId {
			parentId = metadata.IdRootFolder
		}
		metadata := buildMetadata(item.FileId, parentId, item.File)
		if err = d.metaService.Save(parentId, fileId, metadata, !metadata.IsFolder(), false); err != nil {
			return
		}
	}
	return
}

func buildMetadata(id string, parentId string, file *client.File) *metadata.CachedDriveFile {
	lastMod, _ := time.Parse(layoutDateTime, file.ModifiedDate)
	return &metadata.CachedDriveFile{
		Id:          id,
		ParentId:    parentId, // ignoring multiple parents
		Name:        file.Title,
		MimeType:    file.MimeType,
		FileSize:    file.FileSize,
		Md5Checksum: file.Md5Checksum,
		LastMod:     lastMod,
	}
}
