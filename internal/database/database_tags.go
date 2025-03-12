package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/huandu/go-sqlbuilder"
)

// GetTags returns a list of tags from the database.
// If opts.WithBookmarkCount is true, the result will include the number of bookmarks for each tag.
// If opts.BookmarkID is not 0, the result will include only the tags for the specified bookmark.
// If opts.OrderBy is set, the result will be ordered by the specified column.
func (db *dbbase) GetTags(ctx context.Context, opts model.DBListTagsOptions) ([]model.TagDTO, error) {
	sb := db.Flavor().NewSelectBuilder()

	sb.Select("t.id", "t.name")
	sb.From("tag t")

	// Treat the case where we want the bookmark count and filter by bookmark ID as a special case:
	// If we only want one of them, we can use a JOIN and GROUP BY.
	// If we want both, we need to use a subquery to get the count of bookmarks for each tag filtered
	// by bookmark ID.
	if opts.WithBookmarkCount && opts.BookmarkID == 0 {
		// Join with bookmark_tag and group by tag ID to get the count of bookmarks for each tag
		sb.JoinWithOption(sqlbuilder.LeftJoin, "bookmark_tag bt", "bt.tag_id = t.id")
		sb.SelectMore("COUNT(bt.tag_id) AS bookmark_count")
		sb.GroupBy("t.id")
	} else if opts.BookmarkID > 0 {
		// If we want the bookmark count, we need to use a subquery to get the count of bookmarks for each tag
		if opts.WithBookmarkCount {
			sb.SelectMore(
				sb.BuilderAs(
					db.Flavor().NewSelectBuilder().Select("COUNT(bt2.tag_id)").From("bookmark_tag bt2").Where("bt2.tag_id = t.id"),
					"bookmark_count",
				),
			)
		}

		// Join with bookmark_tag and filter by bookmark ID to get the tags for a specific bookmark
		sb.JoinWithOption(sqlbuilder.RightJoin, "bookmark_tag bt",
			sb.And(
				"bt.tag_id = t.id",
				sb.Equal("bt.bookmark_id", opts.BookmarkID),
			),
		)
		sb.Where(sb.IsNotNull("t.id"))
	}

	if opts.OrderBy == model.DBTagOrderByTagName {
		sb.OrderBy("t.name")
	}

	query, args := sb.Build()
	query = db.ReaderDB().Rebind(query)

	slog.Info("GetTags query", "query", query, "args", args)

	var tags []model.TagDTO
	err := db.ReaderDB().SelectContext(ctx, &tags, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	return tags, nil
}
