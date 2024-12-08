package domains

import (
	"context"
	"fmt"

	"github.com/go-shiori/shiori/internal/dependencies"
	"github.com/go-shiori/shiori/internal/model"
)

type TagsDomain struct {
	deps *dependencies.Dependencies
}

func NewTagsDomain(deps *dependencies.Dependencies) model.TagsDomain {
	return &TagsDomain{
		deps: deps,
	}
}

func (d *TagsDomain) GetTags(ctx context.Context) ([]model.TagDTO, error) {
	tags, err := d.deps.Database.GetTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("error retrieving tags: %w", err)
	}

	tagDTOs := make([]model.TagDTO, len(tags))
	for i, tag := range tags {
		tagDTOs[i] = tag.ToDTO()
	}

	return tagDTOs, nil
}

func (d *TagsDomain) CreateTag(ctx context.Context, tag model.TagDTO) (model.TagDTO, error) {
	// Create tag
	err := d.deps.Database.CreateTags(ctx, tag.ToTag())
	if err != nil {
		return model.TagDTO{}, fmt.Errorf("error creating tag: %w", err)
	}

	// Since we can't get the created tag directly,
	// return the input tag as is
	return tag, nil
}

func (d *TagsDomain) CreateTags(ctx context.Context, tags ...model.TagDTO) (model.TagDTO, error) {
	// Convert DTOs to Tags
	modelTags := make([]model.Tag, len(tags))
	for i, tag := range tags {
		modelTags[i] = tag.ToTag()
	}

	// Create tags
	err := d.deps.Database.CreateTags(ctx, modelTags...)
	if err != nil {
		return model.TagDTO{}, fmt.Errorf("error creating tags: %w", err)
	}

	// Return the first tag as per interface definition
	return tags[0], nil
}

func (d *TagsDomain) UpdateTag(ctx context.Context, tag model.TagDTO) (model.TagDTO, error) {
	// Update tag
	err := d.deps.Database.UpdateTag(ctx, tag.ToTag())
	if err != nil {
		return model.TagDTO{}, fmt.Errorf("error updating tag: %w", err)
	}

	return tag, nil
}

func (d *TagsDomain) DeleteTag(ctx context.Context, tagID model.DBID) error {
	// Delete tag
	err := d.deps.Database.DeleteTag(ctx, tagID)
	if err != nil {
		return fmt.Errorf("error deleting tag: %w", err)
	}

	return nil
}
