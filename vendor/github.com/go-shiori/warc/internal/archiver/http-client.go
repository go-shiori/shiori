package archiver

import (
	"crypto/tls"
	"net/http"
	"net/http/cookiejar"
	"time"
)

var httpClient *http.Client

func init() {
	jar, _ := cookiejar.New(nil)
	httpClient = &http.Client{
		Timeout: time.Minute,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Jar: jar,
	}
}
