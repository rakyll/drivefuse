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

package metadata

import (
	"database/sql"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/rakyll/drivefuse/logger"
	"github.com/rakyll/drivefuse/third_party/github.com/coopernurse/gorp"
	_ "github.com/rakyll/drivefuse/third_party/github.com/mattn/go-sqlite3"
)

const (
	OpNone = iota
	OpDownload
	OpUpload
	OpDelete

	MimeTypeFolder = "application/vnd.google-apps.folder"
	IdRoot         = "root"

	keyLargestChangeId = "largest-change-id"
)

// CachedDriveFile represents metadata about a Drive file or folder.
// TODO(burcud): Rename it to FileEntry
type CachedDriveFile struct {
	LocalId       int64
	LocalParentId int64
	Id            string
	Name          string
	LastMod       time.Time
	Md5Checksum   string
	LastEtag      string
	FileSize      int64
	IsDir         bool

	Op int
}

type KeyValueEntry struct {
	Key   string
	Value string
}

// MetaService implements utility methods to retrieve, save, delete
// metadata about Google Drive files/folders.
type MetaService struct {
	dbmap *gorp.DbMap

	mu sync.RWMutex // TODO(burcud): Lock for each file ID indiviually
}

// Initiates a new MetaService.
func New(dbPath string) (metaservice *MetaService, err error) {
	var dbase *sql.DB
	if dbase, err = sql.Open("sqlite3", dbPath); err != nil {
		return
	}
	metaservice = &MetaService{dbmap: &gorp.DbMap{Db: dbase, Dialect: &gorp.SqliteDialect{}}}
	if err = metaservice.setup(); err != nil {
		return
	}
	return metaservice, nil
}

// Permanently saves a file/folder's metadata.
func (m *MetaService) RemoteMod(remoteId string, newParentRemoteId string, data *CachedDriveFile) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	logger.V("Caching metadata for", remoteId)
	var parentFile *CachedDriveFile
	if newParentRemoteId != "" {
		if parentFile, err = m.getByRemoteId(newParentRemoteId); err != nil {
			return err
		}
	}

	var file *CachedDriveFile
	if file, err = m.getByRemoteId(remoteId); err != nil {
		return err
	}
	if file == nil {
		file = &CachedDriveFile{Id: remoteId}
	}
	if data.Md5Checksum != file.Md5Checksum && !data.IsDir {
		file.Op = OpDownload
	}
	file.Id = remoteId

	file.Name = data.Name
	file.LastMod = data.LastMod
	file.Md5Checksum = data.Md5Checksum
	file.LastEtag = data.LastEtag
	file.FileSize = data.FileSize
	file.IsDir = data.IsDir
	file.LocalParentId = 0
	if parentFile != nil {
		file.LocalParentId = parentFile.LocalId
	}
	if file.LocalId > 0 {
		_, err = m.dbmap.Update(file)
		return
	}
	return m.dbmap.Insert(file)
}

func (m *MetaService) RemoteRm(remoteId string) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// TODO: Handle directories recursively
	logger.V("Deleting metadata for", remoteId)
	var file *CachedDriveFile
	if file, err = m.getByRemoteId(remoteId); err != nil {
		return err
	}
	if file == nil {
		return
	}
	file.Op = OpDelete
	_, err = m.dbmap.Update(file)
	return err
}

func (m *MetaService) LocalCreate(localParentId int64, name string, filesize int64, isDir bool) (*CachedDriveFile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	file := &CachedDriveFile{
		LocalParentId: localParentId,
		Name:          name,
		LastMod:       time.Now(),
		FileSize:      filesize,
		IsDir:         isDir,
		Op:            OpUpload,
	}
	err := m.dbmap.Insert(file)
	return file, err
}

func (m *MetaService) LocalMod(localParentId int64, name string, newParentId int64, newName string, newFileSize int64) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var files []*CachedDriveFile
	_, err = m.dbmap.Select(&files, "select * from files where localparentid = :localparentid and name = :name and op != :opdelete", map[string]interface{}{
		"localparentid": localParentId,
		"name":          name,
		"opdelete":      OpDelete,
	})
	if err != nil || len(files) == 0 {
		return err
	}
	file := files[0]
	file.Name = newName
	file.LocalParentId = newParentId
	if newFileSize > -1 {
		file.FileSize = newFileSize
	}
	file.LastMod = time.Now()
	file.Op = OpUpload
	_, err = m.dbmap.Update(file)
	return err
}

func (m *MetaService) LocalRm(localParentId int64, name string, isDir bool) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var files []*CachedDriveFile
	_, err = m.dbmap.Select(&files, "select * from files where localparentid = :localparentid and name = :name and op != :opdelete", map[string]interface{}{
		"localparentid": localParentId,
		"name":          name,
		"opdelete":      OpDelete,
	})
	if err != nil || len(files) == 0 {
		return err
	}
	file := files[0]
	file.Op = OpDelete
	_, err = m.dbmap.Update(file)
	return err
}

func (m *MetaService) ListDownloads(limit int64, min int64, max int64) (files []*CachedDriveFile, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// TODO: order by lastMod
	_, err = m.dbmap.Select(&files, "select * from files where op = :op and filesize >= :min and filesize < :max limit :limit", map[string]interface{}{
		"op":    OpDownload,
		"min":   min,
		"max":   max,
		"limit": limit,
	})
	return files, err
}

// Looks up for files under parentId, named with name.
func (m *MetaService) GetChildrenWithName(localparentid int64, name string) (file *CachedDriveFile, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var files []*CachedDriveFile
	_, err = m.dbmap.Select(&files, "select * from files where localparentid = :localparentid and name = :name and op != :opdelete", map[string]interface{}{
		"localparentid": localparentid,
		"name":          name,
		"opdelete":      OpDelete,
	})
	if err != nil || len(files) == 0 {
		return nil, err
	}
	return files[0], nil
}

// Gets the children of folder identified by parentId.
func (m *MetaService) GetChildren(localparentid int64) (files []*CachedDriveFile, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, err = m.dbmap.Select(&files, "select * from files where localparentid = :localparentid and op != :opdelete", map[string]interface{}{
		"localparentid": localparentid,
		"opdelete":      OpDelete,
	})
	return
}

// Enqueues a file into the upload or download queue.
func (m *MetaService) SetOp(localId int64, op int) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var file *CachedDriveFile
	if file, err = m.getByLocalId(localId); err != nil {
		return err
	}
	file.Op = op
	_, err = m.dbmap.Update(file)
	return err
}

// Gets the largest change id synchnonized.
func (m *MetaService) GetLargestChangeId() (largestId int64, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var val string
	if val, err = m.getKey(keyLargestChangeId); err != nil {
		return
	}
	largestId, err = strconv.ParseInt(val, 0, 64)
	return
}

// Persists the largest change id synchnonized.
func (m *MetaService) SaveLargestChangeId(id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	logger.V("Saving largest change Id", id)
	e := &KeyValueEntry{Key: keyLargestChangeId, Value: fmt.Sprintf("%d", id)}
	val, err := m.getKey(keyLargestChangeId)
	if err != nil {
		return err
	}
	if val == "" {
		return m.dbmap.Insert(e)
	}
	_, err = m.dbmap.Update(e)
	return err
}

// Sets up the sqlite db, creates required tables and indexes.
func (m *MetaService) setup() error {
	m.dbmap.AddTableWithName(CachedDriveFile{}, "files").SetKeys(true, "LocalId")
	m.dbmap.AddTableWithName(KeyValueEntry{}, "info").SetKeys(false, "Key")
	return m.dbmap.CreateTablesIfNotExists()
}

func (m *MetaService) getByRemoteId(remoteId string) (*CachedDriveFile, error) {
	var files []*CachedDriveFile
	_, err := m.dbmap.Select(&files, "select * from files where id = :remoteid", map[string]interface{}{
		"remoteid": remoteId,
	})
	if err != nil || len(files) == 0 {
		return nil, err
	}
	return files[0], err
}

func (m *MetaService) getByLocalId(localId int64) (*CachedDriveFile, error) {
	var files []*CachedDriveFile
	_, err := m.dbmap.Select(&files, "select * from files where localid = :id", map[string]interface{}{
		"id": localId,
	})
	if err != nil || len(files) == 0 {
		return nil, err
	}
	return files[0], err
}

func (m *MetaService) getKey(key string) (value string, err error) {
	var vals []string
	_, err = m.dbmap.Select(&vals, "select value from info where key = ?", key)
	if err != nil || len(vals) == 0 {
		return
	}
	return vals[0], err
}
