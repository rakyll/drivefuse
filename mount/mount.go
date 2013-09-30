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

package mount

import (
	"os"
	"time"

	"github.com/rakyll/drivefuse/blob"
	"github.com/rakyll/drivefuse/fileio"
	"github.com/rakyll/drivefuse/metadata"
	"github.com/rakyll/drivefuse/third_party/code.google.com/p/rsc/fuse"
)

var (
	metaService *metadata.MetaService
	blobManager *blob.Manager
	downloader  *fileio.Downloader
)

type GoogleDriveFS struct{}

func MountAndServe(mountPoint string, meta *metadata.MetaService, blogMngr *blob.Manager, down *fileio.Downloader) error {
	metaService = meta
	blobManager = blogMngr
	downloader = down
	c, err := fuse.Mount(mountPoint)
	if err != nil {
		return err
	}
	c.Serve(GoogleDriveFS{})
	return nil
}

func (GoogleDriveFS) Root() (fuse.Node, fuse.Error) {
	return GoogleDriveFolder{Id: metadata.IdRootFolder}, nil
}

type GoogleDriveFolder struct {
	Id       string
	Name     string
	MimeType string
	Size     int64
	LastMod  time.Time
}

type GoogleDriveFile struct {
	Id          string
	Name        string
	MimeType    string
	Md5Checksum string
	Size        int64
	LastMod     time.Time
}

func (f GoogleDriveFolder) Attr() fuse.Attr {
	return fuse.Attr{
		Mode:  os.ModeDir | 0400,
		Uid:   uint32(os.Getuid()),
		Gid:   uint32(os.Getgid()),
		Mtime: f.LastMod,
	}
}

func (f GoogleDriveFolder) Lookup(name string, intr fuse.Intr) (fuse.Node, fuse.Error) {
	switch name {
	// ignore some MacOSX lookups
	case "._.", ".hidden", ".DS_Store", "mach_kernel", "Backups.backupdb":
		return nil, fuse.ENOENT
	}

	file, err := metaService.LookUp(f.Id, name)
	if err != nil || file == nil {
		return nil, fuse.ENOENT
	}
	if file.MimeType == metadata.MimeTypeFolder {
		return &GoogleDriveFolder{
			Id:      file.Id,
			Name:    file.Name,
			LastMod: file.LastMod}, nil
	}
	return GoogleDriveFile{
		Id:          file.Id,
		Name:        file.Name,
		Size:        file.FileSize,
		Md5Checksum: file.Md5Checksum,
		LastMod:     file.LastMod}, nil
}

func (f GoogleDriveFolder) ReadDir(intr fuse.Intr) ([]fuse.Dirent, fuse.Error) {
	ents := []fuse.Dirent{}
	children, _ := metaService.GetChildren(f.Id)
	for _, item := range children {
		ents = append(ents, fuse.Dirent{Name: item.Name})
	}
	return ents, nil
}

func (f GoogleDriveFile) Attr() fuse.Attr {
	return fuse.Attr{
		Mode:  0400,
		Uid:   uint32(os.Getuid()),
		Gid:   uint32(os.Getgid()),
		Size:  uint64(f.Size),
		Mtime: f.LastMod,
	}
}

func (f GoogleDriveFile) Read(req *fuse.ReadRequest, res *fuse.ReadResponse, intr fuse.Intr) fuse.Error {
	var blob []byte
	var err error

	if blob, _, err = blobManager.Read(f.Id, f.Md5Checksum, req.Offset, req.Size); err != nil {
		// TODO: add a loading icon and etc
		// TODO: force add the file to the download queue
		return nil
	}
	res.Data = blob
	return nil
}

// TODO(burcud): implement mkdir, rename and write
