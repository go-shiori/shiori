package dependencies

import (
	"github.com/go-shiori/shiori/internal/config"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/sirupsen/logrus"
)

type Dependencies struct {
	log      *logrus.Logger
	domains  *domains
	config   *config.Config
	database model.DB
}

func (d *Dependencies) Logger() *logrus.Logger {
	return d.log
}

func (d *Dependencies) Domains() model.DomainDependencies {
	return d.domains
}

func (d *Dependencies) Config() *config.Config {
	return d.config
}

func (d *Dependencies) Database() model.DB {
	return d.database
}

type domains struct {
	auth      model.AuthDomain
	accounts  model.AccountsDomain
	bookmarks model.BookmarksDomain
	archiver  model.ArchiverDomain
	storage   model.StorageDomain
	tags      model.TagsDomain
}

func (d *domains) Auth() model.AuthDomain                       { return d.auth }
func (d *domains) SetAuth(auth model.AuthDomain)                { d.auth = auth }
func (d *domains) Accounts() model.AccountsDomain               { return d.accounts }
func (d *domains) SetAccounts(accounts model.AccountsDomain)    { d.accounts = accounts }
func (d *domains) Bookmarks() model.BookmarksDomain             { return d.bookmarks }
func (d *domains) SetBookmarks(bookmarks model.BookmarksDomain) { d.bookmarks = bookmarks }
func (d *domains) Archiver() model.ArchiverDomain               { return d.archiver }
func (d *domains) SetArchiver(archiver model.ArchiverDomain)    { d.archiver = archiver }
func (d *domains) Storage() model.StorageDomain                 { return d.storage }
func (d *domains) SetStorage(storage model.StorageDomain)       { d.storage = storage }
func (d *domains) Tags() model.TagsDomain                       { return d.tags }
func (d *domains) SetTags(tags model.TagsDomain)                { d.tags = tags }

var _ model.DomainDependencies = (*domains)(nil)

func NewDependencies(log *logrus.Logger, db model.DB, cfg *config.Config) *Dependencies {
	return &Dependencies{
		log:      log,
		config:   cfg,
		database: db,
		domains:  &domains{},
	}
}
