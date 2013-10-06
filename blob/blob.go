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

package blob

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/rakyll/drivefuse/logger"
)

type Manager struct {
	blobPath string
}

func New(blobPath string) *Manager {
	return &Manager{blobPath: blobPath}
}

func (f *Manager) Save(id int64, checksum string, rc io.ReadCloser) error {
	f.cleanup(id, checksum)
	if err := os.MkdirAll(f.getBlobDir(id), 0750); err != nil {
		return err
	}
	file, err := os.OpenFile(f.getBlobPath(id, checksum), os.O_CREATE|os.O_RDWR, 0750)
	if file == nil && err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(rc)
	writer := bufio.NewWriter(file)
	p := make([]byte, 4096)
	for {
		n, err := reader.Read(p)
		if err == io.EOF {
			break
		}
		_, err = writer.Write(p[:n])
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Manager) Read(id int64, checksum string, seek int64, l int) (blob []byte, size int64, err error) {
	var file *os.File
	file, err = os.Open(f.getBlobPath(id, checksum))
	if err != nil {
		return
	}
	defer file.Close()

	blob = make([]byte, l)
	file.Seek(seek, 0)
	var s int
	s, err = file.Read(blob)
	return blob, int64(s), err
}

func (f *Manager) Delete(id int64) error {
	// TODO(burcud): rm directory if not required anymore
	return f.cleanup(id, "*")
}

func (f *Manager) cleanup(id int64, checksum string) (err error) {
	var blobs []os.FileInfo
	if blobs, err = ioutil.ReadDir(f.getBlobDir(id)); err != nil {
		return
	}
	for _, file := range blobs {
		if file.Name() != f.getBlobName(id, checksum) && strings.Contains(file.Name(), f.getBlobName(id, "")) {
			logger.V("Deleting blob", file.Name())
			// errors are not show stoppers here, they will cost additional disk space
			// we can get rid of on the next removal try.
			if rmErr := os.Remove(path.Join(f.getBlobDir(id), file.Name())); rmErr != nil {
				logger.V(rmErr)
			}
		}
	}
	return nil
}

func (f *Manager) getBlobDir(id int64) string {
	idStr := fmt.Sprintf("%d", id)
	return path.Join(f.blobPath, idStr[0:1])
}

func (f *Manager) getBlobName(id int64, checksum string) string {
	return fmt.Sprintf("%d==%s", id, checksum)
}

func (f *Manager) getBlobPath(id int64, checksum string) string {
	return path.Join(f.getBlobDir(id), f.getBlobName(id, checksum))
}
