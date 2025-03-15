package webapp

import (
	"embed"
)

//go:embed dist/index.html
var Templates embed.FS

//go:embed dist/assets dist/*.ico
var Assets embed.FS
