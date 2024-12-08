package domains_test

import (
	"context"
	"testing"

	"github.com/go-shiori/shiori/internal/domains"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/go-shiori/shiori/internal/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestTagsDomainGetTags(t *testing.T) {
	ctx := context.TODO()
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
	domain := domains.NewTagsDomain(deps)

	t.Run("empty tags list", func(t *testing.T) {
		tags, err := domain.GetTags(ctx)
		require.NoError(t, err)
		require.Empty(t, tags)
	})

	t.Run("with tags", func(t *testing.T) {
		// Create some test tags first
		tag1 := model.TagDTO{Name: "test1"}
		tag2 := model.TagDTO{Name: "test2"}

		_, err := domain.CreateTag(ctx, tag1)
		require.NoError(t, err)
		_, err = domain.CreateTag(ctx, tag2)
		require.NoError(t, err)

		// Get all tags
		tags, err := domain.GetTags(ctx)
		require.NoError(t, err)
		require.Len(t, tags, 2)
		require.Contains(t, []string{tags[0].Name, tags[1].Name}, "test1")
		require.Contains(t, []string{tags[0].Name, tags[1].Name}, "test2")
	})
}

func TestTagsDomainCreateTag(t *testing.T) {
	ctx := context.TODO()
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
	domain := domains.NewTagsDomain(deps)

	t.Run("create single tag", func(t *testing.T) {
		tag := model.TagDTO{Name: "test"}
		created, err := domain.CreateTag(ctx, tag)
		require.NoError(t, err)
		require.Equal(t, tag.Name, created.Name)

		// Verify it exists
		tags, err := domain.GetTags(ctx)
		require.NoError(t, err)
		require.Len(t, tags, 1)
		require.Equal(t, tag.Name, tags[0].Name)
	})
}

func TestTagsDomainCreateTags(t *testing.T) {
	ctx := context.TODO()
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
	domain := domains.NewTagsDomain(deps)

	t.Run("create multiple tags", func(t *testing.T) {
		tags := []model.TagDTO{
			{Name: "test1"},
			{Name: "test2"},
			{Name: "test3"},
		}

		created, err := domain.CreateTags(ctx, tags...)
		require.NoError(t, err)
		require.Equal(t, tags[0].Name, created[0].Name)

		// Verify all exist
		allTags, err := domain.GetTags(ctx)
		require.NoError(t, err)
		require.Len(t, allTags, 3)
		tagNames := []string{allTags[0].Name, allTags[1].Name, allTags[2].Name}
		require.Contains(t, tagNames, "test1")
		require.Contains(t, tagNames, "test2")
		require.Contains(t, tagNames, "test3")
	})
}

func TestTagsDomainUpdateTag(t *testing.T) {
	ctx := context.TODO()
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
	domain := domains.NewTagsDomain(deps)

	t.Run("update existing tag", func(t *testing.T) {
		// Create initial tag
		tag := model.TagDTO{Name: "test"}
		created, err := domain.CreateTag(ctx, tag)
		require.NoError(t, err)

		// Update the tag
		updated := model.TagDTO{
			ID:   created.ID,
			Name: "updated",
		}
		result, err := domain.UpdateTag(ctx, updated)
		require.NoError(t, err)
		require.Equal(t, updated.Name, result.Name)

		// Verify the update
		tags, err := domain.GetTags(ctx)
		require.NoError(t, err)
		require.Len(t, tags, 1)
		require.Equal(t, "updated", tags[0].Name)
	})
}

func TestTagsDomainDeleteTag(t *testing.T) {
	ctx := context.TODO()
	logger := logrus.New()
	_, deps := testutil.GetTestConfigurationAndDependencies(t, ctx, logger)
	domain := domains.NewTagsDomain(deps)

	t.Run("delete existing tag", func(t *testing.T) {
		// Create a tag first
		tag := model.TagDTO{Name: "test"}
		created, err := domain.CreateTag(ctx, tag)
		require.NoError(t, err)

		// Delete the tag
		err = domain.DeleteTag(ctx, model.DBID(created.ID))
		require.NoError(t, err)

		// Verify it's gone
		tags, err := domain.GetTags(ctx)
		require.NoError(t, err)
		require.Empty(t, tags)
	})

	t.Run("delete non-existent tag", func(t *testing.T) {
		err := domain.DeleteTag(ctx, model.DBID(999))
		require.Error(t, err)
	})
}
