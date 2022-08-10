package domain

import (
	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/model"
)

type TagsDomain struct {
	db database.DB
}

func (d TagsDomain) GetTags() ([]model.Tag, error) {
	return d.db.GetTags()
}

func NewTagsDomain(db database.DB) TagsDomain {
	return TagsDomain{
		db: db,
	}
}
