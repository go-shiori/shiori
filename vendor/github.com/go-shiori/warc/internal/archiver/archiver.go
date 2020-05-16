package archiver

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/go-shiori/warc/internal/processor"
	"github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"
)

// Request is struct that contains page data that want to be archived.
type Request struct {
	Reader      io.Reader
	URL         string
	ContentType string
}

// Archiver is struct that do the archival.
type Archiver struct {
	sync.RWMutex

	DB         *bbolt.DB
	UserAgent  string
	LogEnabled bool

	resourceMap map[string]struct{}
}

// Start starts the archival process
func (arc *Archiver) Start(req Request) error {
	if arc.resourceMap == nil {
		arc.resourceMap = make(map[string]struct{})
	}

	return arc.archive(req, true)
}

func (arc *Archiver) archive(req Request, root bool) error {
	// Check if this request already processed before
	arc.RLock()
	_, processed := arc.resourceMap[req.URL]
	arc.RUnlock()

	if processed {
		return nil
	}

	// Download page if needed
	if req.Reader == nil || req.ContentType == "" {
		arc.logInfo("Downloading %s\n", req.URL)

		resp, err := arc.downloadPage(req.URL)
		if err != nil {
			return fmt.Errorf("failed to download %s: %v", req.URL, err)
		}
		defer resp.Body.Close()

		req.Reader = resp.Body
		req.ContentType = resp.Header.Get("Content-Type")
	}

	// Process input
	var err error
	resource := processor.Resource{}
	subResources := []processor.Resource{}
	processorRequest := processor.Request{
		Reader: req.Reader,
		URL:    req.URL,
	}

	switch {
	case strings.Contains(req.ContentType, "text/html"):
		resource, subResources, err = processor.ProcessHTMLFile(processorRequest)
		if !root && !resource.IsEmbed {
			subResources = []processor.Resource{}
		}
	case strings.Contains(req.ContentType, "text/css") && !root:
		resource, subResources, err = processor.ProcessCSSFile(processorRequest)
	default:
		resource, err = processor.ProcessGeneralFile(processorRequest)
	}

	if err != nil {
		return fmt.Errorf("failed to archive %s: %v", req.URL, err)
	}

	// Save resource to storage
	if root {
		resource.Name = "archive-root"
	}

	err = arc.saveResource(resource, req.ContentType)
	if err != nil {
		return fmt.Errorf("failed to save %s: %v", req.URL, err)
	}

	// Save this resource to map
	arc.Lock()
	arc.resourceMap[req.URL] = struct{}{}
	arc.Unlock()

	arc.logInfo("Saved %s (%d)\n", resource.URL, len(resource.Content))

	// Archive the sub resources
	wg := sync.WaitGroup{}
	wg.Add(len(subResources))

	semaphore := make(chan struct{}, 5)
	defer close(semaphore)

	for _, subResource := range subResources {
		go func(subResource processor.Resource) {
			// Make sure to finish the WG
			defer wg.Done()

			// Register goroutine to semaphore
			semaphore <- struct{}{}
			defer func() {
				<-semaphore
			}()

			// Archive the sub resource
			var subResContent io.Reader
			if len(subResource.Content) > 0 {
				subResContent = bytes.NewBuffer(subResource.Content)
			}

			subResRequest := Request{
				Reader: subResContent,
				URL:    subResource.URL,
			}

			err := arc.archive(subResRequest, false)
			if err != nil {
				arc.logWarning("Failed to save %s: %v\n", subResource.URL, err)
			}
		}(subResource)
	}

	wg.Wait()

	return nil
}

// DownloadData downloads data from the specified URL.
func (arc *Archiver) downloadPage(url string) (*http.Response, error) {
	// Prepare request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Send request
	req.Header.Set("User-Agent", arc.UserAgent)
	return httpClient.Do(req)
}

func (arc *Archiver) saveResource(resource processor.Resource, contentType string) error {
	// Compress content
	buffer := bytes.NewBuffer(nil)
	gzipper := gzip.NewWriter(buffer)

	_, err := gzipper.Write(resource.Content)
	if err != nil {
		return fmt.Errorf("compress failed: %v", err)
	}

	err = gzipper.Close()
	if err != nil {
		return fmt.Errorf("compress failed: %v", err)
	}

	err = arc.DB.Batch(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(resource.Name))
		if bucket != nil {
			return nil
		}

		bucket, err := tx.CreateBucketIfNotExists([]byte(resource.Name))
		if err != nil {
			return err
		}

		err = bucket.Put([]byte("content"), buffer.Bytes())
		if err != nil {
			return err
		}

		err = bucket.Put([]byte("type"), []byte(contentType))
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (arc *Archiver) logInfo(format string, args ...interface{}) {
	if arc.LogEnabled {
		logrus.Infof(format, args...)
	}
}

func (arc *Archiver) logWarning(format string, args ...interface{}) {
	if arc.LogEnabled {
		logrus.Warnf(format, args...)
	}
}
