package model

// Variables set my the main package coming from ldflags
var (
	BuildVersion = "dev"
	BuildCommit  = "none"
	BuildDate    = "unknown"
)

const (
	// ShioriNamespace
	ShioriURLNamespace = "https://github.com/go-shiori/shiori"

)

var (
	// UserAgent: default user agent used for downloads and archival
	UserAgent = "Shiori/" + BuildVersion + " (" + BuildCommit + ") (+https://github.com/go-shiori/shiori)"
)
