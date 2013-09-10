package syncer

import (
	"log"
	"sync"
	"time"

	"metadata"
	client "third_party/code.google.com/p/google-api-go-client/drive/v2"
)

const (
	intervalSync = 30 * time.Second // TODO: should be adaptive
)

type CachedSyncer struct {
	remoteService *client.Service
	metaService   *metadata.MetaService

	mu sync.RWMutex
}

func New(service *client.Service, metaService *metadata.MetaService) *CachedSyncer {
	return &CachedSyncer{
		remoteService: service,
		metaService:   metaService,
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

	log.Println("Started syncer...")
	err = d.syncInbound(isForce)
	if err != nil {
		log.Println("error during sync", err)
	}
	log.Println("Done syncing...")
	return
}

func (d *CachedSyncer) syncOutbound(rootId string, isRecursive bool, isForce bool) error {
	panic("not implemented")
	return nil
}

func (d *CachedSyncer) syncInbound(isForce bool) (err error) {
	var largestChangeId int64
	largestChangeId, err = d.metaService.GetLargestChangeId()
	if isForce || err != nil {
		largestChangeId = 0
	} else {
		largestChangeId += 1
	}
	isInitialSync := largestChangeId == 0

	// retrieve metadata about root
	var rootFile *client.File
	if rootFile, err = d.remoteService.Files.Get(metadata.IdRootFolder).Do(); err != nil {
		return
	}

	data := &metadata.CachedDriveFile{
		Id:          metadata.IdRootFolder,
		ParentId:    "",
		Name:        rootFile.Title,
		MimeType:    rootFile.MimeType,
		FileSize:    rootFile.FileSize,
		Md5Checksum: rootFile.Md5Checksum,
		LastMod:     time.Now(), // TODO: parse
	}

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
	log.Println("merging changes starting with pageToken:", pageToken, "and startChangeId", startChangeId)

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
		// delete
		if d.metaService.Delete(item.FileId); err != nil {
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
		metadata := &metadata.CachedDriveFile{
			Id:          item.FileId,
			ParentId:    parentId,        // ignore multiple parents
			Name:        item.File.Title, // TODO: rename duplicates
			MimeType:    item.File.MimeType,
			FileSize:    item.File.FileSize,
			Md5Checksum: item.File.Md5Checksum,
			LastMod:     time.Now(), // TODO: parse
		}
		if err = d.metaService.Save(parentId, fileId, metadata, !metadata.IsFolder(), false); err != nil {
			return
		}
	}
	return
}
