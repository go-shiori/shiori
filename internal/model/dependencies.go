package model

import (
	"github.com/go-shiori/shiori/internal/config"
	"github.com/sirupsen/logrus"
)

// Dependencies represents the interface for application dependencies
type Dependencies interface {
	Logger() *logrus.Logger
	Domains() DomainDependencies
	Config() *config.Config
	Database() DB
}

// DomainDependencies represents the interface for domain-specific dependencies
type DomainDependencies interface {
	Auth() AuthDomain
	SetAuth(auth AuthDomain)
	Accounts() AccountsDomain
	SetAccounts(accounts AccountsDomain)
	Bookmarks() BookmarksDomain
	SetBookmarks(bookmarks BookmarksDomain)
	Archiver() ArchiverDomain
	SetArchiver(archiver ArchiverDomain)
	Storage() StorageDomain
	SetStorage(storage StorageDomain)
	Tags() TagsDomain
	SetTags(tags TagsDomain)
}
