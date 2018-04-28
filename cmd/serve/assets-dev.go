// +build dev

package serve

import "net/http"

var assets = http.Dir("view")
