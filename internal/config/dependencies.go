package config

import (
	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/domains"
	"github.com/sirupsen/logrus"
)

type Dependencies struct {
	Log      *logrus.Logger
	Database database.DB
	Config   *Config
	Domains  struct {
		Auth     domains.AccountsDomain
		Archiver domains.ArchiverDomain
	}
}

func NewDependencies(log *logrus.Logger, db database.DB, cfg *Config) *Dependencies {
	return &Dependencies{
		Log:      log,
		Config:   cfg,
		Database: db,
	}
}
