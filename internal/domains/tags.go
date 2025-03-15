package domains

import (
	"context"
	"errors"

	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/model"
)

type tagsDomain struct {
	deps model.Dependencies
}

func NewTagsDomain(deps model.Dependencies) model.TagsDomain {
	return &tagsDomain{deps: deps}
}

func (d *tagsDomain) ListTags(ctx context.Context, opts model.ListTagsOptions) ([]model.TagDTO, error) {
	tags, err := d.deps.Database().GetTags(ctx, model.DBListTagsOptions(opts))
	if err != nil {
		return nil, err
	}

	return tags, nil
}

func (d *tagsDomain) CreateTag(ctx context.Context, tagDTO model.TagDTO) (model.TagDTO, error) {
	tag := tagDTO.ToTag()
	createdTag, err := d.deps.Database().CreateTag(ctx, tag)
	if err != nil {
		return model.TagDTO{}, err
	}

	return createdTag.ToDTO(), nil
}

func (d *tagsDomain) GetTag(ctx context.Context, id int) (model.TagDTO, error) {
	tag, exists, err := d.deps.Database().GetTag(ctx, id)
	if err != nil {
		return model.TagDTO{}, err
	}
	if !exists {
		return model.TagDTO{}, model.ErrNotFound
	}
	return tag, nil
}

func (d *tagsDomain) UpdateTag(ctx context.Context, tagDTO model.TagDTO) (model.TagDTO, error) {
	tag := tagDTO.ToTag()
	err := d.deps.Database().UpdateTag(ctx, tag)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return model.TagDTO{}, model.ErrNotFound
		}
		return model.TagDTO{}, err
	}

	// Fetch the updated tag to return
	updatedTag, err := d.GetTag(ctx, tag.ID)
	if err != nil {
		return model.TagDTO{}, err
	}

	return updatedTag, nil
}

func (d *tagsDomain) DeleteTag(ctx context.Context, id int) error {
	if err := d.deps.Database().DeleteTag(ctx, id); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return model.ErrNotFound
		}
		return err
	}

	return nil
}

// TagExists checks if a tag with the given ID exists
func (d *tagsDomain) TagExists(ctx context.Context, id int) (bool, error) {
	return d.deps.Database().TagExists(ctx, id)
}
