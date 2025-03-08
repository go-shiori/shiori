package domains

import (
	"context"

	"github.com/go-shiori/shiori/internal/model"
)

type tagsDomain struct {
	deps model.Dependencies
}

func NewTagsDomain(deps model.Dependencies) model.TagsDomain {
	return &tagsDomain{deps: deps}
}

func (d *tagsDomain) ListTags(ctx context.Context) ([]model.TagDTO, error) {
	tags, err := d.deps.Database().GetTags(ctx)
	if err != nil {
		return nil, err
	}

	return tags, nil
}

func (d *tagsDomain) CreateTag(ctx context.Context, tagDTO model.TagDTO) (model.TagDTO, error) {
	tag := tagDTO.ToTag()
	err := d.deps.Database().CreateTags(ctx, tag)
	if err != nil {
		return model.TagDTO{}, err
	}
	return tag.ToDTO(), nil
}
