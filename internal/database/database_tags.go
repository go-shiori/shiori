package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jmoiron/sqlx"
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

	// Add search condition if search term is provided
	if opts.Search != "" {
		// Note: Search and BookmarkID filtering are mutually exclusive as per requirements
		sb.Where(sb.Like("t.name", "%"+opts.Search+"%"))
	}

	if opts.OrderBy == model.DBTagOrderByTagName {
		sb.OrderBy("t.name")
	}

	query, args := sb.Build()
	query = db.ReaderDB().Rebind(query)

	tags := []model.TagDTO{}
	err := db.ReaderDB().SelectContext(ctx, &tags, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	return tags, nil
}

// AddTagToBookmark adds a tag to a bookmark
func (db *dbbase) AddTagToBookmark(ctx context.Context, bookmarkID int, tagID int) error {
	// Insert the bookmark-tag association
	insertSb := db.Flavor().NewInsertBuilder()
	insertSb.InsertInto("bookmark_tag")
	insertSb.Cols("bookmark_id", "tag_id")
	insertSb.Values(bookmarkID, tagID)

	insertQuery, insertArgs := insertSb.Build()
	insertQuery = db.WriterDB().Rebind(insertQuery)

	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		// First check if the association already exists using sqlbuilder
		selectSb := db.Flavor().NewSelectBuilder()
		selectSb.Select("1")
		selectSb.From("bookmark_tag")
		selectSb.Where(
			selectSb.And(
				selectSb.Equal("bookmark_id", bookmarkID),
				selectSb.Equal("tag_id", tagID),
			),
		)

		selectQuery, selectArgs := selectSb.Build()
		selectQuery = db.ReaderDB().Rebind(selectQuery)

		var exists int
		err := tx.QueryRowContext(ctx, selectQuery, selectArgs...).Scan(&exists)

		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to check if tag is already associated: %w", err)
		}

		// If it doesn't exist, insert it
		if err == sql.ErrNoRows {
			_, err = tx.ExecContext(ctx, insertQuery, insertArgs...)
			if err != nil {
				return fmt.Errorf("failed to add tag to bookmark: %w", err)
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// RemoveTagFromBookmark removes a tag from a bookmark
func (db *dbbase) RemoveTagFromBookmark(ctx context.Context, bookmarkID int, tagID int) error {
	// Delete the bookmark-tag association
	deleteSb := db.Flavor().NewDeleteBuilder()
	deleteSb.DeleteFrom("bookmark_tag")
	deleteSb.Where(
		deleteSb.And(
			deleteSb.Equal("bookmark_id", bookmarkID),
			deleteSb.Equal("tag_id", tagID),
		),
	)

	query, args := deleteSb.Build()
	query = db.WriterDB().Rebind(query)

	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to remove tag from bookmark: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

// TagExists checks if a tag with the given ID exists in the database
func (db *dbbase) TagExists(ctx context.Context, tagID int) (bool, error) {
	sb := db.Flavor().NewSelectBuilder()
	sb.Select("1")
	sb.From("tag")
	sb.Where(sb.Equal("id", tagID))
	sb.Limit(1)

	query, args := sb.Build()
	query = db.ReaderDB().Rebind(query)

	var exists int
	err := db.ReaderDB().QueryRowContext(ctx, query, args...).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if tag exists: %w", err)
	}

	return true, nil
}

// BookmarkExists checks if a bookmark with the given ID exists in the database
func (db *dbbase) BookmarkExists(ctx context.Context, bookmarkID int) (bool, error) {
	sb := db.Flavor().NewSelectBuilder()
	sb.Select("1")
	sb.From("bookmark")
	sb.Where(sb.Equal("id", bookmarkID))
	sb.Limit(1)

	query, args := sb.Build()
	query = db.ReaderDB().Rebind(query)

	var exists int
	err := db.ReaderDB().QueryRowContext(ctx, query, args...).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if bookmark exists: %w", err)
	}

	return true, nil
}
