package config

import (
	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/domains"
	"go.uber.org/zap"
)

type Dependencies struct {
	Log      *zap.Logger
	Database database.DB
	Domains  struct {
		Auth domains.AuthDomain
	}
}

func NewDependencies(log *zap.Logger, db database.DB) *Dependencies {
	return &Dependencies{
		Log:      log,
		Database: db,
	}
}
