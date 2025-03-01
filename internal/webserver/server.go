package webserver

import (
	"time"

	"github.com/go-shiori/shiori/internal/model"
	cch "github.com/patrickmn/go-cache"
)

// Config is parameter that used for starting web server
type Config struct {
	DB            model.DB
	DataDir       string
	ServerAddress string
	ServerPort    int
	RootPath      string
	Log           bool
}

// GetLegacyHandler returns a legacy handler to use with the new webserver
func GetLegacyHandler(cfg Config, dependencies model.Dependencies) *Handler {
	return &Handler{
		DB:        cfg.DB,
		DataDir:   cfg.DataDir,
		UserCache: cch.New(time.Hour, 10*time.Minute),
		// SessionCache: cch.New(time.Hour, 10*time.Minute),
		ArchiveCache: cch.New(time.Minute, 5*time.Minute),
		RootPath:     cfg.RootPath,
		Log:          cfg.Log,
		dependencies: dependencies,
	}
}
