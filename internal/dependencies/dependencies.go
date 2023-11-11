package dependencies

import (
	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/domains"
	"github.com/sirupsen/logrus"
)

type Dependencies struct {
	Log      *logrus.Logger
	Database database.DB
	Config   *config.Config
	Domains  struct {
		Auth      domains.AccountsDomain
		Archiver  domains.ArchiverDomain
		Bookmarks domains.BookmarksDomain
	}
}

func NewDependencies(log *logrus.Logger, db database.DB, cfg *config.Config) *Dependencies {
	return &Dependencies{
		Log:      log,
		Config:   cfg,
		Database: db,
	}
}
