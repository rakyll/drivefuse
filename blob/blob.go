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
	"io"
	"os"
	"path"
)

type Manager struct {
	blobPath string
}

func New(blobPath string) *Manager {
	return &Manager{blobPath: blobPath}
}

func (f *Manager) Save(id string, checksum string, rc io.ReadCloser) error {
	// TODO(burcud): Remove old versions of the same file
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

func (f *Manager) Read(id string, checksum string, seek int64, l int) (blob []byte, size int64, err error) {
	var file *os.File
	file, err = os.Open(f.getBlobPath(id, checksum))
	if err != nil {
		return
	}
	blob = make([]byte, l)
	file.Seek(seek, 0)
	var s int
	s, err = file.Read(blob)
	return blob, int64(s), err
}

func (f *Manager) getBlobPath(id string, checksum string) string {
	// TODO: shard the files, fs perf issue here
	return path.Join(f.blobPath, id+"=="+checksum)
}
