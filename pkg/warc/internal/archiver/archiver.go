package archiver

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.etcd.io/bbolt"
)

// Archiver is struct for archiving an URL and its resources.
type Archiver struct {
	sync.RWMutex
	sync.WaitGroup

	DB          *bbolt.DB
	ChDone      chan struct{}
	ChErrors    chan error
	ChWarnings  chan error
	ChRequest   chan ResourceURL
	ResourceMap map[string]struct{}
	LogEnabled  bool
}

// Close closes channels that used by the Archiver.
func (arc *Archiver) Close() {
	close(arc.ChErrors)
	close(arc.ChWarnings)
	close(arc.ChRequest)
}

// StartArchiver starts the archival process.
func (arc *Archiver) StartArchiver() []error {
	go func() {
		time.Sleep(time.Second)
		arc.Wait()
		close(arc.ChDone)
	}()

	// Download the URL concurrently.
	// After download finished, parse response to extract resources
	// URL inside it. After that, send it to channel to download again.
	errors := make([]error, 0)
	warnings := make([]error, 0)

	func() {
		for {
			select {
			case <-arc.ChDone:
				return
			case err := <-arc.ChErrors:
				errors = append(errors, err)
			case err := <-arc.ChWarnings:
				warnings = append(warnings, err)
			case res := <-arc.ChRequest:
				arc.RLock()
				_, exist := arc.ResourceMap[res.DownloadURL]
				arc.RUnlock()

				if !exist {
					arc.Add(1)
					go arc.archive(res)
				}
			}
		}
	}()

	// Print log message if required
	if arc.LogEnabled {
		nErrors := len(errors)
		nWarnings := len(warnings)
		arc.Logf(infoLog, "Download finished with %d warnings and %d errors\n", nWarnings, nErrors)

		if nWarnings > 0 {
			fmt.Println()
			for _, warning := range warnings {
				arc.Log(warningLog, warning)
			}
		}

		if nErrors > 0 {
			for _, err := range errors {
				arc.Log(errorLog, err)
			}
		}
	}

	return nil
}

// archive downloads a subresource and save it to storage.
func (arc *Archiver) archive(res ResourceURL) {
	// Make sure to decrease wait group once finished
	defer arc.Done()

	// Download resource
	resp, err := DownloadData(res.DownloadURL)
	if err != nil {
		arc.ChErrors <- fmt.Errorf("failed to download %s: %v", res.DownloadURL, err)
		return
	}
	defer resp.Body.Close()

	// Process resource depending on its type.
	// Since this `archive` method only used for processing sub
	// resource, we will only process the CSS and HTML sub resources.
	// For other file, we will simply download it as it is.
	var result ProcessResult
	var subResources []ResourceURL
	cType := resp.Header.Get("Content-Type")

	switch {
	case strings.Contains(cType, "text/html") && res.IsEmbedded:
		result, subResources, err = arc.ProcessHTMLFile(res, resp.Body)
	case strings.Contains(cType, "text/css"):
		result, subResources, err = arc.ProcessCSSFile(res, resp.Body)
	default:
		result, err = arc.ProcessOtherFile(res, resp.Body)
	}

	if err != nil {
		arc.ChErrors <- fmt.Errorf("failed to process %s: %v", res.DownloadURL, err)
		return
	}

	// Add this url to resource map
	arc.Lock()
	arc.ResourceMap[res.DownloadURL] = struct{}{}
	arc.Unlock()

	// Save content to storage
	arc.Logf(infoLog, "Downloaded %s\n"+
		"\tArchive name %s\n"+
		"\tParent %s\n"+
		"\tSize %d Bytes\n",
		res.DownloadURL,
		res.ArchivalURL,
		res.Parent,
		resp.ContentLength)

	result.ContentType = cType
	err = arc.SaveToStorage(result)
	if err != nil {
		arc.ChErrors <- fmt.Errorf("failed to save %s: %v", res.DownloadURL, err)
		return
	}

	// Send sub resource to request channel
	for _, subRes := range subResources {
		arc.ChRequest <- subRes
	}
}

// SaveToStorage save processing result to storage.
func (arc *Archiver) SaveToStorage(result ProcessResult) error {
	// Compress content
	buffer := bytes.NewBuffer(nil)
	gzipper := gzip.NewWriter(buffer)

	_, err := gzipper.Write(result.Content)
	if err != nil {
		return fmt.Errorf("compress failed: %v", err)
	}

	err = gzipper.Close()
	if err != nil {
		return fmt.Errorf("compress failed: %v", err)
	}

	err = arc.DB.Batch(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(result.Name))
		if bucket != nil {
			return nil
		}

		bucket, err := tx.CreateBucketIfNotExists([]byte(result.Name))
		if err != nil {
			return err
		}

		err = bucket.Put([]byte("content"), buffer.Bytes())
		if err != nil {
			return err
		}

		err = bucket.Put([]byte("type"), []byte(result.ContentType))
		if err != nil {
			return err
		}

		return nil
	})

	return err
}
