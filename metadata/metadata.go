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
	"log"
	"strconv"
	"sync"
	"time"

	_ "third_party/github.com/mattn/go-sqlite3"
)

const (
	MimeTypeFolder = "application/vnd.google-apps.folder"
	IdRootFolder   = "root"

	keyStarted         = "started-before"
	keyLargestChangeId = "largest-change-id"
)

// CachedDriveFile represents metadata about a Drive file or folder.
// TODO(burcud): Rename it to Metadata
type CachedDriveFile struct {
	Id          string
	ParentId    string
	Name        string
	MimeType    string
	LastMod     time.Time
	Md5Checksum string
	FileSize    int64
}

// Returns true if the object is a folder.
func (file *CachedDriveFile) IsFolder() bool {
	return file.MimeType == MimeTypeFolder
}

// MetaService implements utility methods to retrieve, save, delete
// metadata about Google Drive files/folders.
type MetaService struct {
	db *sql.DB

	mu sync.RWMutex // TODO(burcud): Lock for each file ID indiviually
}

// Initiates a new MetaService.
func New(dbPath string) (metaservice *MetaService, err error) {
	var dbase *sql.DB
	if dbase, err = sql.Open("sqlite3", dbPath); err != nil {
		return
	}
	metaservice = &MetaService{db: dbase}
	if err = metaservice.setup(); err != nil {
		return
	}
	return metaservice, nil
}

// Cleans up and closes resources used by the meta service.
func (m *MetaService) Close() error {
	return m.db.Close()
}

// Permanently saves a file/folder's metadata.
func (m *MetaService) Save(parentId string, id string, data *CachedDriveFile, download bool, upload bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Println("Caching metadata for", id)
	return m.upsertFile(data, download, upload)
}

func (m *MetaService) Delete(id string) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Println("Deleting metadata for", id)
	return m.deleteFile(id)
}

func (m *MetaService) ListDownloads(limit int64) ([]*CachedDriveFile, error) {
	// TODO: order by lastMod
	return m.listFiles(fmt.Sprintf(sqlListDownloads, limit))
}

// Looks up for files under parentId, named with name.
func (m *MetaService) LookUp(parentId string, name string) (file *CachedDriveFile, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	query := fmt.Sprintf(sqlLookup, parentId, name)
	var files []*CachedDriveFile
	if files, err = m.listFiles(query); err != nil {
		return
	}
	if len(files) > 0 {
		file = files[0]
	}
	return
}

// Gets the children of folder identified by parentId.
func (m *MetaService) GetChildren(parentId string) (output []*CachedDriveFile, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	query := fmt.Sprintf(sqlChildren, parentId)
	return m.listFiles(query)
}

func (m *MetaService) InitFile(id string) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, err = m.db.Exec(sqlSetInited, id)
	return
}

// Enqueues a file into the upload or download queue.
func (m *MetaService) EnqueueForIO(queueName string, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.updateIOQueue(queueName, id, 1)
}

// Removes the file from upload or download queue.
func (m *MetaService) DequeueFromIO(queueName string, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.updateIOQueue(queueName, id, 0)
}

// Gets the largest change id synchnonized.
func (m *MetaService) GetLargestChangeId() (largestId int64, err error) {
	var val string
	val, err = m.getValue(keyLargestChangeId)
	if err != nil {
		return
	}
	largestId, err = strconv.ParseInt(val, 0, 64)
	if err != nil {
		return
	}
	return
}

// Persists the largest change id synchnonized.
func (m *MetaService) SaveLargestChangeId(id int64) error {
	return m.setValue(keyLargestChangeId, fmt.Sprintf("%d", id))
}
