package archiver

import (
	"crypto/tls"
	"net/http"
	"net/http/cookiejar"
	"time"
)

var (
	defaultClient *http.Client
)

func init() {
	jar, _ := cookiejar.New(nil)
	defaultClient = &http.Client{
		Timeout: time.Minute,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Jar: jar,
	}
}

// DownloadData downloads data from the specified URL.
func DownloadData(url string) (*http.Response, error) {
	// Prepare request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Send request
	req.Header.Set("User-Agent", "Shiori/2.0.0 (+https://github.com/go-shiori/shiori)")
	return defaultClient.Do(req)
}
