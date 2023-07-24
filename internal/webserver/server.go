package webserver

import (
	"time"

	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/database"
	cch "github.com/patrickmn/go-cache"
)

// Config is parameter that used for starting web server
type Config struct {
	DB            database.DB
	DataDir       string
	ServerAddress string
	ServerPort    int
	RootPath      string
	Log           bool
}

func GetLegacyHandler(cfg Config, dependencies *config.Dependencies) *Handler {
	return &Handler{
		DB:           cfg.DB,
		DataDir:      cfg.DataDir,
		UserCache:    cch.New(time.Hour, 10*time.Minute),
		SessionCache: cch.New(time.Hour, 10*time.Minute),
		ArchiveCache: cch.New(time.Minute, 5*time.Minute),
		RootPath:     cfg.RootPath,
		Log:          cfg.Log,
		depenencies:  dependencies,
	}
}
