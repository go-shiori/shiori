//go:build !dev
// +build !dev

package webserver

import (
	"github.com/go-shiori/shiori/internal"
	"io/fs"
)

var assets fs.FS

func init() {
	var err error
	assets, err = fs.Sub(internal.Assets, "view")
	if err != nil {
		panic(err)
	}
}
