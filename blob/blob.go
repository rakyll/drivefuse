package blob

import (
	"bufio"
	"fmt"
	"io"
	"os"
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
	return fmt.Sprintf("%s%c%s==%s", f.blobPath, os.PathSeparator, id, checksum)
}
