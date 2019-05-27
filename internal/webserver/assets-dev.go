// +build dev

package webserver

import (
	"net/http"
)

var assets = http.Dir("internal/view")

func init() {
	developmentMode = true
}
