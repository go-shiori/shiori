package dependencies

import (
	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/sirupsen/logrus"
)

type Domains struct {
	Archiver  model.ArchiverDomain
	Auth      model.AccountsDomain
	Bookmarks model.BookmarksDomain
	Storage   model.StorageDomain
}

type Dependencies struct {
	Log      *logrus.Logger
	Database database.DB
	Config   *config.Config
	Domains  *Domains
}

func NewDependencies(log *logrus.Logger, db database.DB, cfg *config.Config) *Dependencies {
	return &Dependencies{
		Log:      log,
		Config:   cfg,
		Database: db,
		Domains:  &Domains{},
	}
}
