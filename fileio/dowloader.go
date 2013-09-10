package fileio

import (
	"log"
	"net/http"
	"sync"
	"time"

	"blob"
	"metadata"
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
	log.Println("Downloading", id, checksum)
	var (
		resp *http.Response
		err  error
	)
	if resp, err = d.client.Get(BaseUrlDownloadHost + "/" + id); err != nil {
		log.Println("error downloading", id, err)
		return
	}

	if resp.StatusCode == 404 {
		d.metaService.DequeueFromIO("download", id)
		log.Println("error downloading [not found]", id)
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Println("error downloading [not ok]", id, resp.StatusCode)
		return
	}

	defer resp.Body.Close()
	err = d.blobMngr.Save(id, checksum, resp.Body)
	if err != nil {
		log.Println(err)
		return
	}

	err = d.metaService.InitFile(id)
	if err != nil {
		log.Println(err)
		return
	}

	d.metaService.DequeueFromIO("download", id)
}
