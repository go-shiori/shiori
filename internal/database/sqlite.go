package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"slices"

	_ "modernc.org/sqlite"
)

var sqliteMigrations = []migration{
	newFileMigration("0.0.0", "0.1.0", "sqlite/0000_system"),
	newFileMigration("0.1.0", "0.2.0", "sqlite/0001_initial"),
	newFuncMigration("0.2.0", "0.3.0", func(db *sql.DB) error {
		// Ensure that bookmark table has `has_content` column and account table has `config` column
		// for users upgrading from <1.5.4 directly into this version.
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}
		defer tx.Rollback()

		_, err = tx.Exec(`ALTER TABLE bookmark ADD COLUMN has_content BOOLEAN DEFAULT FALSE NOT NULL`)
		if err != nil && strings.Contains(err.Error(), `duplicate column name`) {
			tx.Rollback()
		} else if err != nil {
			return fmt.Errorf("failed to add has_content column to bookmark table: %w", err)
		} else if err == nil {
			if errCommit := tx.Commit(); errCommit != nil {
				return fmt.Errorf("failed to commit transaction: %w", errCommit)
			}
		}

		tx, err = db.Begin()
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}
		defer tx.Rollback()

		_, err = tx.Exec(`ALTER TABLE account ADD COLUMN config JSON NOT NULL DEFAULT '{}'`)
		if err != nil && strings.Contains(err.Error(), `duplicate column name`) {
			tx.Rollback()
		} else if err != nil {
			return fmt.Errorf("failed to add config column to account table: %w", err)
		} else if err == nil {
			if errCommit := tx.Commit(); errCommit != nil {
				return fmt.Errorf("failed to commit transaction: %w", errCommit)
			}
		}

		return nil
	}),
	newFileMigration("0.3.0", "0.4.0", "sqlite/0002_denormalize_content"),
	newFileMigration("0.4.0", "0.5.0", "sqlite/0003_uniq_id"),
	newFileMigration("0.5.0", "0.6.0", "sqlite/0004_created_time"),
}

// SQLiteDatabase is implementation of Database interface
// for connecting to SQLite3 database.
type SQLiteDatabase struct {
	dbbase
}

// withTx executes the given function within a transaction.
// If the function returns an error, the transaction is rolled back.
// Otherwise, the transaction is committed.
func (db *SQLiteDatabase) withTx(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	tx, err := db.writer.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Will be a no-op if tx.Commit() is called

	if err := fn(tx); err != nil {
		// Return the error directly without wrapping
		return fmt.Errorf("transaction failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// withTxRetry executes the given function within a transaction with retry logic.
// It will retry up to 3 times if the database is locked, with exponential backoff.
// For other errors, it returns immediately.
func (db *SQLiteDatabase) withTxRetry(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	maxRetries := 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		err := db.withTx(ctx, fn)
		if err == nil {
			return nil
		}

		if strings.Contains(err.Error(), "database is locked") {
			lastErr = err
			time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
			continue
		}

		return fmt.Errorf("transaction failed after retry: %w", err)
	}

	return fmt.Errorf("transaction failed after max retries, last error: %w", lastErr)
}

// Init sets up the SQLite database with optimal settings for both reader and writer connections
func (db *SQLiteDatabase) Init(ctx context.Context) error {
	// Initialize both connections with appropriate settings
	for _, conn := range []*sqlx.DB{db.WriterDB(), db.ReaderDB()} {
		// Reuse connections for up to one hour
		conn.SetConnMaxLifetime(time.Hour)

		// Enable WAL mode for better concurrency
		if _, err := conn.ExecContext(ctx, `PRAGMA journal_mode=WAL`); err != nil {
			return fmt.Errorf("failed to set journal mode: %w", err)
		}

		// Set busy timeout to avoid "database is locked" errors
		if _, err := conn.ExecContext(ctx, `PRAGMA busy_timeout=5000`); err != nil {
			return fmt.Errorf("failed to set busy timeout: %w", err)
		}

		// Other performance and reliability settings
		pragmas := []string{
			`PRAGMA synchronous=NORMAL`,
			`PRAGMA cache_size=-2000`, // Use 2MB of memory for cache
			`PRAGMA foreign_keys=ON`,
		}

		for _, pragma := range pragmas {
			if _, err := conn.ExecContext(ctx, pragma); err != nil {
				return fmt.Errorf("failed to set pragma %s: %w", pragma, err)
			}
		}
	}

	// Use a single connection on the writer to avoid database is locked errors
	db.writer.SetMaxOpenConns(1)

	// Set maximum idle connections for the reader to number of CPUs (maxing at 4)
	db.reader.SetMaxIdleConns(max(4, runtime.NumCPU()))

	return nil
}

type bookmarkContent struct {
	ID      int    `db:"docid"`
	Content string `db:"content"`
	HTML    string `db:"html"`
}

// DBX returns the underlying sqlx.DB object for writes
func (db *SQLiteDatabase) WriterDB() *sqlx.DB {
	return db.dbbase.WriterDB()
}

// ReaderDBx returns the underlying sqlx.DB object for reading
func (db *SQLiteDatabase) ReaderDB() *sqlx.DB {
	return db.dbbase.ReaderDB()
}

// Migrate runs migrations for this database engine
func (db *SQLiteDatabase) Migrate(ctx context.Context) error {
	if err := runMigrations(ctx, db, sqliteMigrations); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// GetDatabaseSchemaVersion fetches the current migrations version of the database
func (db *SQLiteDatabase) GetDatabaseSchemaVersion(ctx context.Context) (string, error) {
	var version string

	err := db.reader.GetContext(ctx, &version, "SELECT database_schema_version FROM shiori_system")
	if err != nil {
		return "", fmt.Errorf("failed to get database schema version: %w", err)
	}

	return version, nil
}

// SetDatabaseSchemaVersion sets the current migrations version of the database
func (db *SQLiteDatabase) SetDatabaseSchemaVersion(ctx context.Context, version string) error {
	if err := db.withTxRetry(ctx, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(ctx, "UPDATE shiori_system SET database_schema_version = ?", version)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to set database schema version: %w", err)
	}

	return nil
}

// SaveBookmarks saves new or updated bookmarks to database.
// Returns the saved ID and error message if any happened.
func (db *SQLiteDatabase) SaveBookmarks(ctx context.Context, create bool, bookmarks ...model.BookmarkDTO) ([]model.BookmarkDTO, error) {
	var result []model.BookmarkDTO

	if err := db.withTxRetry(ctx, func(tx *sqlx.Tx) error {
		// Prepare statement

		stmtInsertBook, err := tx.PreparexContext(ctx, `INSERT INTO bookmark
			(url, title, excerpt, author, public, modified_at, has_content, created_at)
			VALUES(?, ?, ?, ?, ?, ?, ?, ?) RETURNING id`)
		if err != nil {
			return fmt.Errorf("failed to prepare insert book statement: %w", err)
		}

		stmtUpdateBook, err := tx.PreparexContext(ctx, `UPDATE bookmark SET
			url = ?, title = ?,	excerpt = ?, author = ?,
			public = ?, modified_at = ?, has_content = ?
			WHERE id = ?`)
		if err != nil {
			return fmt.Errorf("failed to prepare update book statement: %w", err)
		}

		stmtInsertBookContent, err := tx.PreparexContext(ctx, `INSERT OR REPLACE INTO bookmark_content
			(docid, title, content, html)
			VALUES (?, ?, ?, ?)`)
		if err != nil {
			return fmt.Errorf("failed to prepare insert book content statement: %w", err)
		}

		stmtUpdateBookContent, err := tx.PreparexContext(ctx, `UPDATE bookmark_content SET
			title = ?, content = ?, html = ?
			WHERE docid = ?`)
		if err != nil {
			return fmt.Errorf("failed to prepare update book content statement: %w", err)
		}

		stmtGetTag, err := tx.PreparexContext(ctx, `SELECT id FROM tag WHERE name = ?`)
		if err != nil {
			return fmt.Errorf("failed to prepare get tag statement: %w", err)
		}

		stmtInsertTag, err := tx.PreparexContext(ctx, `INSERT INTO tag (name) VALUES (?)`)
		if err != nil {
			return fmt.Errorf("failed to prepare insert tag statement: %w", err)
		}

		stmtInsertBookTag, err := tx.PreparexContext(ctx, `INSERT OR IGNORE INTO bookmark_tag
			(tag_id, bookmark_id) VALUES (?, ?)`)
		if err != nil {
			return fmt.Errorf("failed to prepare insert book tag statement: %w", err)
		}

		stmtDeleteBookTag, err := tx.PreparexContext(ctx, `DELETE FROM bookmark_tag
			WHERE bookmark_id = ? AND tag_id = ?`)
		if err != nil {
			return fmt.Errorf("failed to execute delete statement: %w", err)
		}

		// Prepare modified time
		modifiedTime := time.Now().UTC().Format(model.DatabaseDateFormat)

		// Execute statements

		for _, book := range bookmarks {
			// Check URL and title
			if book.URL == "" {
				return errors.New("URL must not be empty")
			}

			if book.Title == "" {
				return errors.New("title must not be empty")
			}

			// Set modified time
			if book.ModifiedAt == "" {
				book.ModifiedAt = modifiedTime
			}

			hasContent := book.Content != ""

			// Create or update bookmark
			var err error
			if create {
				book.CreatedAt = modifiedTime
				err = stmtInsertBook.QueryRowContext(ctx,
					book.URL, book.Title, book.Excerpt, book.Author, book.Public, book.ModifiedAt, hasContent, book.CreatedAt).Scan(&book.ID)
			} else {
				_, err = stmtUpdateBook.ExecContext(ctx,
					book.URL, book.Title, book.Excerpt, book.Author, book.Public, book.ModifiedAt, hasContent, book.ID)
			}
			if err != nil {
				return fmt.Errorf("failed to delete bookmark content: %w", err)
			}

			// Try to update it first to check for existence, we can't do an UPSERT here because
			// bookmant_content is a virtual table
			res, err := stmtUpdateBookContent.ExecContext(ctx, book.Title, book.Content, book.HTML, book.ID)
			if err != nil {
				return fmt.Errorf("failed to delete bookmark tag: %w", err)
			}

			rows, err := res.RowsAffected()
			if err != nil {
				return fmt.Errorf("failed to delete bookmark: %w", err)
			}

			if rows == 0 {
				_, err = stmtInsertBookContent.ExecContext(ctx, book.ID, book.Title, book.Content, book.HTML)
				if err != nil {
					return fmt.Errorf("failed to execute delete bookmark tag statement: %w", err)
				}
			}

			// Save book tags
			newTags := []model.TagDTO{}
			for _, tag := range book.Tags {
				t := tag.ToDTO()
				// If it's deleted tag, delete and continue
				if t.Deleted {
					_, err = stmtDeleteBookTag.ExecContext(ctx, book.ID, tag.ID)
					if err != nil {
						return fmt.Errorf("failed to execute delete bookmark statement: %w", err)
					}
					continue
				}

				// Normalize tag name
				tagName := strings.ToLower(tag.Name)
				tagName = strings.Join(strings.Fields(tagName), " ")

				// If tag doesn't have any ID, fetch it from database
				if tag.ID == 0 {
					if err := stmtGetTag.GetContext(ctx, &tag.ID, tagName); err != nil && err != sql.ErrNoRows {
						return fmt.Errorf("failed to get tag ID: %w", err)
					}

					// If tag doesn't exist in database, save it
					if tag.ID == 0 {
						res, err := stmtInsertTag.ExecContext(ctx, tagName)
						if err != nil {
							return fmt.Errorf("failed to get last insert ID for tag: %w", err)
						}

						tagID64, err := res.LastInsertId()
						if err != nil && err != sql.ErrNoRows {
							return fmt.Errorf("failed to insert bookmark tag: %w", err)
						}

						tag.ID = int(tagID64)
						t.ID = int(tagID64)
					}

					if _, err := stmtInsertBookTag.ExecContext(ctx, tag.ID, book.ID); err != nil {
						return fmt.Errorf("failed to execute bookmark tag statement: %w", err)
					}
				}

				newTags = append(newTags, t)
			}

			book.Tags = newTags
			result = append(result, book)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to execute select query for bookmark content: %w", err)
	}

	return result, nil
}

// GetBookmarks fetch list of bookmarks based on submitted options.
func (db *SQLiteDatabase) GetBookmarks(ctx context.Context, opts model.DBGetBookmarksOptions) ([]model.BookmarkDTO, error) {
	// Create initial query
	query := `SELECT
		b.id,
		b.url,
		b.title,
		b.excerpt,
		b.author,
		b.public,
		b.created_at,
		b.modified_at,
		b.has_content
		FROM bookmark b
		WHERE 1`

	// Add where clause
	args := []interface{}{}

	// Add where clause for IDs
	if len(opts.IDs) > 0 {
		query += ` AND b.id IN (?)`
		args = append(args, opts.IDs)
	}

	// Add where clause for search keyword
	if opts.Keyword != "" {
		query += ` AND (b.url LIKE '%' || ? || '%' OR b.excerpt LIKE '%' || ? || '%' OR b.id IN (
			SELECT docid id
			FROM bookmark_content
			WHERE title MATCH ? OR content MATCH ?))`

		args = append(args, opts.Keyword, opts.Keyword)

		// Replace dash with spaces since FTS5 uses `-name` as column identifier and double quote
		// since FTS5 uses double quote as string identifier
		// Reference: https://sqlite.org/fts5.html#fts5_strings
		ftsKeyword := strings.ReplaceAll(opts.Keyword, "-", " ")

		// Properly set double quotes for string literals in sqlite's fts
		ftsKeyword = strings.ReplaceAll(ftsKeyword, "\"", "\"\"")

		args = append(args, "\""+ftsKeyword+"\"", "\""+ftsKeyword+"\"")
	}

	// Add where clause for tags.
	// First we check for * in excluded and included tags,
	// which means all tags will be excluded and included, respectively.
	excludeAllTags := false
	if slices.Contains(opts.ExcludedTags, "*") {
		excludeAllTags = true
		opts.ExcludedTags = []string{}
	}

	includeAllTags := false
	if slices.Contains(opts.Tags, "*") {
		includeAllTags = true
		opts.Tags = []string{}
	}

	// If all tags excluded, we will only show bookmark without tags.
	// In other hand, if all tags included, we will only show bookmark with tags.
	if excludeAllTags {
		query += ` AND b.id NOT IN (SELECT DISTINCT bookmark_id FROM bookmark_tag)`
	} else if includeAllTags {
		query += ` AND b.id IN (SELECT DISTINCT bookmark_id FROM bookmark_tag)`
	}

	// Now we only need to find the normal tags
	if len(opts.Tags) > 0 {
		query += ` AND b.id IN (
			SELECT bt.bookmark_id
			FROM bookmark_tag bt
			LEFT JOIN tag t ON bt.tag_id = t.id
			WHERE t.name IN(?)
			GROUP BY bt.bookmark_id
			HAVING COUNT(bt.bookmark_id) = ?)`

		args = append(args, opts.Tags, len(opts.Tags))
	}

	if len(opts.ExcludedTags) > 0 {
		query += ` AND b.id NOT IN (
			SELECT DISTINCT bt.bookmark_id
			FROM bookmark_tag bt
			LEFT JOIN tag t ON bt.tag_id = t.id
			WHERE t.name IN(?))`

		args = append(args, opts.ExcludedTags)
	}

	// Add order clause
	switch opts.OrderMethod {
	case model.ByLastAdded:
		query += ` ORDER BY b.id DESC`
	case model.ByLastModified:
		query += ` ORDER BY b.modified_at DESC`
	default:
		query += ` ORDER BY b.id`
	}

	if opts.Limit > 0 && opts.Offset >= 0 {
		query += ` LIMIT ? OFFSET ?`
		args = append(args, opts.Limit, opts.Offset)
	}

	// Expand query, because some of the args might be an array
	query, args, err := sqlx.In(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute select query for tags: %w", err)
	}

	// Fetch bookmarks
	bookmarks := []model.BookmarkDTO{}
	err = db.reader.SelectContext(ctx, &bookmarks, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to fetch accounts: %w", err)
	}

	// store bookmark IDs for further enrichment
	var bookmarkIds = make([]int, 0, len(bookmarks))
	for _, book := range bookmarks {
		bookmarkIds = append(bookmarkIds, book.ID)
	}

	if len(bookmarkIds) == 0 {
		return bookmarks, nil
	}

	// If content needed, fetch it separately
	// It's faster than join with virtual table
	if opts.WithContent {
		contents := make([]bookmarkContent, 0, len(bookmarks))
		contentMap := make(map[int]bookmarkContent, len(bookmarks))

		contentQuery, args, err := sqlx.In(`SELECT docid, content, html FROM bookmark_content WHERE docid IN (?)`, bookmarkIds)
		contentQuery = db.reader.Rebind(contentQuery)
		if err != nil {
			return nil, fmt.Errorf("failed to expand tags query with IN clause: %w", err)
		}

		err = db.reader.Select(&contents, contentQuery, args...)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to get tags: %w", err)
		}
		for _, content := range contents {
			contentMap[content.ID] = content
		}
		for i := range bookmarks[:] {
			book := &bookmarks[i]
			if bookmarkContent, found := contentMap[book.ID]; found {
				book.Content = bookmarkContent.Content
				book.HTML = bookmarkContent.HTML
			} else {
				log.Printf("not found content for bookmark %d, but it should be; check DB consistency", book.ID)
			}
		}
	}

	// Fetch tags for each bookmark
	for i, book := range bookmarks {
		tags, err := db.getTagsForBookmark(ctx, book.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get tags: %w", err)
		}
		bookmarks[i].Tags = tags
	}

	return bookmarks, nil
}

func (db *SQLiteDatabase) getTagsForBookmark(ctx context.Context, bookmarkID int) ([]model.TagDTO, error) {
	sb := sqlbuilder.SQLite.NewSelectBuilder()
	sb.Select("t.id", "t.name")
	sb.From("bookmark_tag bt")
	sb.JoinWithOption(sqlbuilder.LeftJoin, "tag t", "bt.tag_id = t.id")
	sb.Where(sb.Equal("bt.bookmark_id", bookmarkID))
	sb.OrderBy("t.name")

	query, args := sb.Build()
	query = db.ReaderDB().Rebind(query)

	tags := []model.TagDTO{}
	err := db.ReaderDB().SelectContext(ctx, &tags, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	return tags, nil
}

// GetBookmarksCount fetch count of bookmarks based on submitted options.
func (db *SQLiteDatabase) GetBookmarksCount(ctx context.Context, opts model.DBGetBookmarksOptions) (int, error) {
	// Create initial query
	query := `SELECT COUNT(b.id)
		FROM bookmark b
		WHERE 1`

	// Add where clause
	args := []interface{}{}

	// Add where clause for IDs
	if len(opts.IDs) > 0 {
		query += ` AND b.id IN (?)`
		args = append(args, opts.IDs)
	}

	// Add where clause for search keyword
	if opts.Keyword != "" {
		query += ` AND (b.url LIKE '%' || ? || '%' OR b.excerpt LIKE '%' || ? || '%' OR b.id IN (
			SELECT docid id
			FROM bookmark_content
			WHERE title MATCH ? OR content MATCH ?))`

		args = append(args, opts.Keyword, opts.Keyword)

		// Replace dash with spaces since FTS5 uses `-name` as column identifier and double quote
		// since FTS5 uses double quote as string identifier
		// Reference: https://sqlite.org/fts5.html#fts5_strings
		ftsKeyword := strings.ReplaceAll(opts.Keyword, "-", " ")

		// Properly set double quotes for string literals in sqlite's fts
		ftsKeyword = strings.ReplaceAll(ftsKeyword, "\"", "\"\"")

		args = append(args, "\""+ftsKeyword+"\"", "\""+ftsKeyword+"\"")
	}

	// Add where clause for tags.
	// First we check for * in excluded and included tags,
	// which means all tags will be excluded and included, respectively.
	excludeAllTags := false
	for _, excludedTag := range opts.ExcludedTags {
		if excludedTag == "*" {
			excludeAllTags = true
			opts.ExcludedTags = []string{}
			break
		}
	}

	includeAllTags := false
	for _, includedTag := range opts.Tags {
		if includedTag == "*" {
			includeAllTags = true
			opts.Tags = []string{}
			break
		}
	}

	// If all tags excluded, we will only show bookmark without tags.
	// In other hand, if all tags included, we will only show bookmark with tags.
	if excludeAllTags {
		query += ` AND b.id NOT IN (SELECT DISTINCT bookmark_id FROM bookmark_tag)`
	} else if includeAllTags {
		query += ` AND b.id IN (SELECT DISTINCT bookmark_id FROM bookmark_tag)`
	}

	// Now we only need to find the normal tags
	if len(opts.Tags) > 0 {
		query += ` AND b.id IN (
			SELECT bt.bookmark_id
			FROM bookmark_tag bt
			LEFT JOIN tag t ON bt.tag_id = t.id
			WHERE t.name IN(?)
			GROUP BY bt.bookmark_id
			HAVING COUNT(bt.bookmark_id) = ?)`

		args = append(args, opts.Tags, len(opts.Tags))
	}

	if len(opts.ExcludedTags) > 0 {
		query += ` AND b.id NOT IN (
			SELECT DISTINCT bt.bookmark_id
			FROM bookmark_tag bt
			LEFT JOIN tag t ON bt.tag_id = t.id
			WHERE t.name IN(?))`

		args = append(args, opts.ExcludedTags)
	}

	// Expand query, because some of the args might be an array
	query, args, err := sqlx.In(query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to expand query with IN clause: %w", err)
	}

	// Fetch count
	var nBookmarks int
	err = db.reader.GetContext(ctx, &nBookmarks, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to get bookmark count: %w", err)
	}

	return nBookmarks, nil
}

// DeleteBookmarks removes all record with matching ids from database.
func (db *SQLiteDatabase) DeleteBookmarks(ctx context.Context, ids ...int) error {
	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		// Prepare queries
		delBookmark := `DELETE FROM bookmark`
		delBookmarkTag := `DELETE FROM bookmark_tag`
		delBookmarkContent := `DELETE FROM bookmark_content`

		// Delete bookmark(s)
		if len(ids) == 0 {
			_, err := tx.ExecContext(ctx, delBookmarkContent)
			if err != nil {
				return fmt.Errorf("failed to prepare delete statement: %w", err)
			}

			_, err = tx.ExecContext(ctx, delBookmarkTag)
			if err != nil {
				return fmt.Errorf("failed to execute delete account statement: %w", err)
			}

			_, err = tx.ExecContext(ctx, delBookmark)
			if err != nil {
				return fmt.Errorf("failed to execute delete bookmark statement: %w", err)
			}
		} else {
			delBookmark += ` WHERE id = ?`
			delBookmarkTag += ` WHERE bookmark_id = ?`
			delBookmarkContent += ` WHERE docid = ?`

			stmtDelBookmark, err := tx.Preparex(delBookmark)
			if err != nil {
				return fmt.Errorf("failed to get bookmark: %w", err)
			}

			stmtDelBookmarkTag, err := tx.Preparex(delBookmarkTag)
			if err != nil {
				return fmt.Errorf("failed to expand query with IN clause: %w", err)
			}

			stmtDelBookmarkContent, err := tx.Preparex(delBookmarkContent)
			if err != nil {
				return fmt.Errorf("failed to delete bookmark content: %w", err)
			}

			for _, id := range ids {
				_, err = stmtDelBookmarkContent.ExecContext(ctx, id)
				if err != nil {
					return fmt.Errorf("failed to delete bookmark: %w", err)
				}

				_, err = stmtDelBookmarkTag.ExecContext(ctx, id)
				if err != nil {
					return fmt.Errorf("failed to delete bookmark tag: %w", err)
				}

				_, err = stmtDelBookmark.ExecContext(ctx, id)
				if err != nil {
					return fmt.Errorf("failed to delete bookmark: %w", err)
				}
			}
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to update database schema version: %w", err)
	}

	return nil
}

// GetBookmark fetches bookmark based on its ID or URL.
// Returns the bookmark and boolean whether it's exist or not.
func (db *SQLiteDatabase) GetBookmark(ctx context.Context, id int, url string) (model.BookmarkDTO, bool, error) {
	// Create the main query builder for bookmark data
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(
		"b.id", "b.url", "b.title", "b.excerpt", "b.author", "b.public", "b.modified_at",
		"bc.content", "bc.html", "b.has_content", "b.created_at")
	sb.From("bookmark b")
	sb.JoinWithOption(sqlbuilder.LeftJoin, "bookmark_content bc", "bc.docid = b.id")

	// Add conditions
	if id != 0 {
		sb.Where(sb.Equal("b.id", id))
	} else if url != "" {
		sb.Where(sb.Equal("b.url", url))
	} else {
		return model.BookmarkDTO{}, false, fmt.Errorf("id or url is required")
	}

	// Build the query
	query, args := sb.Build()
	query = db.ReaderDB().Rebind(query)

	// Execute the query
	book := model.BookmarkDTO{}
	err := db.ReaderDB().GetContext(ctx, &book, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return book, false, nil
		}
		return book, false, fmt.Errorf("failed to get bookmark: %w", err)
	}

	// If bookmark exists, fetch its tags
	if book.ID != 0 {
		// Create query builder for tags
		tagSb := sqlbuilder.NewSelectBuilder()
		tagSb.Select("t.id", "t.name")
		tagSb.From("tag t")
		tagSb.JoinWithOption(sqlbuilder.InnerJoin, "bookmark_tag bt", "bt.tag_id = t.id")
		tagSb.Where(tagSb.Equal("bt.bookmark_id", book.ID))

		// Build the query
		tagQuery, tagArgs := tagSb.Build()
		tagQuery = db.ReaderDB().Rebind(tagQuery)

		// Execute the query
		tags := []model.TagDTO{}
		err = db.ReaderDB().SelectContext(ctx, &tags, tagQuery, tagArgs...)
		if err != nil && err != sql.ErrNoRows {
			return book, false, fmt.Errorf("failed to get bookmark tags: %w", err)
		}

		book.Tags = tags
	}

	return book, true, nil
}

// CreateAccount saves new account to database. Returns error if any happened.
func (db *SQLiteDatabase) CreateAccount(ctx context.Context, account model.Account) (*model.Account, error) {
	var accountID int64
	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		// Check if username already exists
		var exists bool
		err := tx.GetContext(ctx, &exists,
			"SELECT EXISTS(SELECT 1 FROM account WHERE username = ?)",
			account.Username)
		if err != nil {
			return fmt.Errorf("error checking username existence: %w", err)
		}
		if exists {
			return ErrAlreadyExists
		}

		// Insert new account
		query, err := tx.PrepareContext(ctx, `INSERT INTO account
			(username, password, owner, config) VALUES (?, ?, ?, ?)
			RETURNING id`)
		if err != nil {
			return fmt.Errorf("error preparing query: %w", err)
		}

		err = query.QueryRowContext(ctx,
			account.Username, account.Password, account.Owner, account.Config).Scan(&accountID)
		if err != nil {
			return fmt.Errorf("error executing query: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("error running transaction: %w", err)
	}

	account.ID = model.DBID(accountID)
	return &account, nil
}

// UpdateAccount updates account in database.
func (db *SQLiteDatabase) UpdateAccount(ctx context.Context, account model.Account) error {
	if account.ID == 0 {
		return ErrNotFound
	}

	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		// Check if username already exists
		var exists bool
		err := tx.GetContext(ctx, &exists,
			"SELECT EXISTS(SELECT 1 FROM account WHERE username = ? AND id != ?)",
			account.Username, account.ID)
		if err != nil {
			return fmt.Errorf("error checking username existence: %w", err)
		}
		if exists {
			return ErrAlreadyExists
		}

		result, err := tx.ExecContext(ctx, `UPDATE account
			SET username = ?, password = ?, owner = ?, config = ?
			WHERE id = ?`,
			account.Username, account.Password, account.Owner, account.Config, account.ID)
		if err != nil {
			return fmt.Errorf("error updating account: %w", err)
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("error getting rows affected: %w", err)
		}
		if rows == 0 {
			return ErrNotFound
		}

		return nil
	}); err != nil {
		return fmt.Errorf("error running transaction: %w", err)
	}

	return nil
}

// ListAccounts fetch list of account (without its password) based on submitted options.
func (db *SQLiteDatabase) ListAccounts(ctx context.Context, opts model.DBListAccountsOptions) ([]model.Account, error) {
	// Create query
	args := []interface{}{}
	fields := []string{"id", "username", "owner", "config"}
	if opts.WithPassword {
		fields = append(fields, "password")
	}

	query := fmt.Sprintf(`SELECT %s FROM account WHERE 1`, strings.Join(fields, ", "))

	if opts.Keyword != "" {
		query += " AND username LIKE ?"
		args = append(args, "%"+opts.Keyword+"%")
	}

	if opts.Username != "" {
		query += " AND username = ?"
		args = append(args, opts.Username)
	}

	if opts.Owner {
		query += " AND owner = 1"
	}

	// Fetch list account
	accounts := []model.Account{}
	err := db.reader.SelectContext(ctx, &accounts, query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to execute select query: %w", err)
	}

	return accounts, nil
}

// GetAccount fetch account with matching ID.
// Returns the account and boolean whether it's exist or not.
func (db *SQLiteDatabase) GetAccount(ctx context.Context, id model.DBID) (*model.Account, bool, error) {
	account := model.Account{}
	err := db.ReaderDB().GetContext(ctx, &account, `SELECT
		id, username, password, owner, config FROM account WHERE id = ?`,
		id,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return &account, false, ErrNotFound
		}
		return &account, false, fmt.Errorf("error getting account: %w", err)
	}

	return &account, true, nil
}

// DeleteAccount removes record with matching ID.
func (db *SQLiteDatabase) DeleteAccount(ctx context.Context, id model.DBID) error {
	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		result, err := tx.ExecContext(ctx, `DELETE FROM account WHERE id = ?`, id)
		if err != nil {
			return fmt.Errorf("error deleting account: %w", err)
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("error getting rows affected: %w", err)
		}

		if rows == 0 {
			return ErrNotFound
		}

		return nil
	}); err != nil {
		return fmt.Errorf("error running transaction: %w", err)
	}

	return nil
}

// CreateTags creates new tags from submitted objects.
func (db *SQLiteDatabase) CreateTags(ctx context.Context, tags ...model.Tag) ([]model.Tag, error) {
	if len(tags) == 0 {
		return []model.Tag{}, nil
	}

	// Create a slice to hold the created tags
	createdTags := make([]model.Tag, len(tags))
	copy(createdTags, tags)

	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		// For SQLite, we need to insert tags one by one to get their IDs
		stmtInsertTag, err := tx.PrepareContext(ctx, "INSERT INTO tag (name) VALUES (?)")
		if err != nil {
			return fmt.Errorf("failed to prepare tag insertion statement: %w", err)
		}
		defer stmtInsertTag.Close()

		// Insert each tag and get its ID
		for i, tag := range createdTags {
			result, err := stmtInsertTag.ExecContext(ctx, tag.Name)
			if err != nil {
				return fmt.Errorf("failed to insert tag: %w", err)
			}

			// Get the last inserted ID
			tagID, err := result.LastInsertId()
			if err != nil {
				return fmt.Errorf("failed to get last insert ID: %w", err)
			}

			createdTags[i].ID = int(tagID)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to run tag creation transaction: %w", err)
	}

	return createdTags, nil
}

// CreateTag creates a new tag in database.
func (db *SQLiteDatabase) CreateTag(ctx context.Context, tag model.Tag) (model.Tag, error) {
	// Use CreateTags to implement this method
	createdTags, err := db.CreateTags(ctx, tag)
	if err != nil {
		return model.Tag{}, err
	}

	if len(createdTags) == 0 {
		return model.Tag{}, fmt.Errorf("failed to create tag")
	}

	return createdTags[0], nil
}

// RenameTag change the name of a tag.
func (db *SQLiteDatabase) RenameTag(ctx context.Context, id int, newName string) error {
	sb := sqlbuilder.NewUpdateBuilder()
	sb.Update("tag")
	sb.Set(sb.Assign("name", newName))
	sb.Where(sb.Equal("id", id))

	query, args := sb.Build()
	query = db.WriterDB().Rebind(query)

	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to rename tag: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

// GetTag fetch a tag by its ID.
func (db *SQLiteDatabase) GetTag(ctx context.Context, id int) (model.TagDTO, bool, error) {
	sb := sqlbuilder.SQLite.NewSelectBuilder()
	sb.Select("t.id", "t.name", "COUNT(bt.tag_id) bookmark_count")
	sb.From("tag t")
	sb.JoinWithOption(sqlbuilder.LeftJoin, "bookmark_tag bt", "bt.tag_id = t.id")
	sb.Where(sb.Equal("t.id", id))
	sb.GroupBy("t.id")
	sb.OrderBy("t.name")

	query, args := sb.Build()
	query = db.ReaderDB().Rebind(query)

	var tag model.TagDTO
	err := db.ReaderDB().GetContext(ctx, &tag, query, args...)
	if err == sql.ErrNoRows {
		return model.TagDTO{}, false, nil
	}
	if err != nil {
		return model.TagDTO{}, false, fmt.Errorf("failed to get tag: %w", err)
	}

	return tag, true, nil
}

// UpdateTag updates a tag in the database.
func (db *SQLiteDatabase) UpdateTag(ctx context.Context, tag model.Tag) error {
	sb := sqlbuilder.NewUpdateBuilder()
	sb.Update("tag")
	sb.Set(sb.Assign("name", tag.Name))
	sb.Where(sb.Equal("id", tag.ID))

	query, args := sb.Build()
	query = db.WriterDB().Rebind(query)

	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to update tag: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

// DeleteTag removes a tag from the database.
func (db *SQLiteDatabase) DeleteTag(ctx context.Context, id int) error {
	// First, check if the tag exists
	_, exists, err := db.GetTag(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check if tag exists: %w", err)
	}
	if !exists {
		return ErrNotFound
	}

	// Delete all bookmark_tag associations
	deleteAssocSb := sqlbuilder.NewDeleteBuilder()
	deleteAssocSb.DeleteFrom("bookmark_tag")
	deleteAssocSb.Where(deleteAssocSb.Equal("tag_id", id))

	deleteAssocQuery, deleteAssocArgs := deleteAssocSb.Build()
	deleteAssocQuery = db.WriterDB().Rebind(deleteAssocQuery)

	// Then, delete the tag itself
	deleteTagSb := sqlbuilder.NewDeleteBuilder()
	deleteTagSb.DeleteFrom("tag")
	deleteTagSb.Where(deleteTagSb.Equal("id", id))

	deleteTagQuery, deleteTagArgs := deleteTagSb.Build()
	deleteTagQuery = db.WriterDB().Rebind(deleteTagQuery)

	if err := db.withTx(ctx, func(tx *sqlx.Tx) error {
		// Delete bookmark_tag associations
		_, err := tx.ExecContext(ctx, deleteAssocQuery, deleteAssocArgs...)
		if err != nil {
			return fmt.Errorf("failed to delete tag associations: %w", err)
		}

		// Delete the tag
		_, err = tx.ExecContext(ctx, deleteTagQuery, deleteTagArgs...)
		if err != nil {
			return fmt.Errorf("failed to delete tag: %w", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// SaveBookmark saves a single bookmark to database without handling tags.
// It only updates the bookmark data in the database.
func (db *SQLiteDatabase) SaveBookmark(ctx context.Context, bookmark model.Bookmark) error {
	if bookmark.ID <= 0 {
		return fmt.Errorf("bookmark ID must be greater than 0")
	}

	bookmark.ModifiedAt = time.Now().UTC().Format(model.DatabaseDateFormat)

	// Check URL and title
	if bookmark.URL == "" {
		return errors.New("URL must not be empty")
	}

	if bookmark.Title == "" {
		return errors.New("title must not be empty")
	}

	// Use sqlbuilder to build the update query
	sb := sqlbuilder.NewUpdateBuilder()
	sb.Update("bookmark")
	sb.Set(
		sb.Assign("url", bookmark.URL),
		sb.Assign("title", bookmark.Title),
		sb.Assign("excerpt", bookmark.Excerpt),
		sb.Assign("author", bookmark.Author),
		sb.Assign("public", bookmark.Public),
		sb.Assign("modified_at", bookmark.ModifiedAt),
		sb.Assign("has_content", bookmark.HasContent),
	)
	sb.Where(sb.Equal("id", bookmark.ID))

	query, args := sb.Build()
	query = db.WriterDB().Rebind(query)

	return db.withTx(ctx, func(tx *sqlx.Tx) error {
		// Update bookmark
		_, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to update bookmark: %w", err)
		}

		return nil
	})
}

// BulkUpdateBookmarkTags updates tags for multiple bookmarks.
// It ensures that all bookmarks and tags exist before proceeding.
func (db *SQLiteDatabase) BulkUpdateBookmarkTags(ctx context.Context, bookmarkIDs []int, tagIDs []int) error {
	if len(bookmarkIDs) == 0 || len(tagIDs) == 0 {
		return nil
	}

	// Convert int slices to any slices for sqlbuilder
	bookmarkIDsIface := make([]any, len(bookmarkIDs))
	for i, id := range bookmarkIDs {
		bookmarkIDsIface[i] = id
	}

	// Verify all bookmarks exist
	bookmarkSb := sqlbuilder.NewSelectBuilder()
	bookmarkSb.Select("id")
	bookmarkSb.From("bookmark")
	bookmarkSb.Where(bookmarkSb.In("id", bookmarkIDsIface...))

	bookmarkQuery, bookmarkArgs := bookmarkSb.Build()
	bookmarkQuery = db.ReaderDB().Rebind(bookmarkQuery)

	var existingBookmarkIDs []int
	err := db.ReaderDB().SelectContext(ctx, &existingBookmarkIDs, bookmarkQuery, bookmarkArgs...)
	if err != nil {
		return fmt.Errorf("failed to check bookmarks: %w", err)
	}

	if len(existingBookmarkIDs) != len(bookmarkIDs) {
		// Find which bookmarks don't exist
		missingBookmarkIDs := model.SliceDifference(bookmarkIDs, existingBookmarkIDs)
		return fmt.Errorf("some bookmarks do not exist: %v", missingBookmarkIDs)
	}

	tagIDsIface := make([]any, len(tagIDs))
	for i, id := range tagIDs {
		tagIDsIface[i] = id
	}

	// Verify all tags exist
	tagSb := sqlbuilder.NewSelectBuilder()
	tagSb.Select("id")
	tagSb.From("tag")
	tagSb.Where(tagSb.In("id", tagIDsIface...))

	tagQuery, tagArgs := tagSb.Build()
	tagQuery = db.ReaderDB().Rebind(tagQuery)

	var existingTagIDs []int
	err = db.ReaderDB().SelectContext(ctx, &existingTagIDs, tagQuery, tagArgs...)
	if err != nil {
		return fmt.Errorf("failed to check tags: %w", err)
	}

	if len(existingTagIDs) != len(tagIDs) {
		// Find which tags don't exist
		missingTagIDs := model.SliceDifference(tagIDs, existingTagIDs)
		return fmt.Errorf("some tags do not exist: %v", missingTagIDs)
	}

	return db.withTx(ctx, func(tx *sqlx.Tx) error {
		// Delete existing bookmark-tag associations
		deleteSb := sqlbuilder.NewDeleteBuilder()
		deleteSb.DeleteFrom("bookmark_tag")
		deleteSb.Where(deleteSb.In("bookmark_id", bookmarkIDsIface...))

		deleteQuery, deleteArgs := deleteSb.Build()
		deleteQuery = tx.Rebind(deleteQuery)

		_, err := tx.ExecContext(ctx, deleteQuery, deleteArgs...)
		if err != nil {
			return fmt.Errorf("failed to delete existing bookmark tags: %w", err)
		}

		// Insert new bookmark-tag associations
		if len(tagIDs) > 0 {
			// Build insert statement for bookmark tags
			insertSb := sqlbuilder.NewInsertBuilder()
			// SQLite syntax for INSERT OR IGNORE
			insertSb.SQL("INSERT OR IGNORE INTO")
			insertSb.SQL("bookmark_tag")
			insertSb.SQL("(bookmark_id, tag_id)")
			insertSb.SQL("VALUES (?, ?)")

			insertQuery := insertSb.String()
			insertQuery = tx.Rebind(insertQuery)

			stmtInsertBookTag, err := tx.PreparexContext(ctx, insertQuery)
			if err != nil {
				return fmt.Errorf("failed to prepare insert book tag statement: %w", err)
			}
			defer stmtInsertBookTag.Close()

			// Insert new tags
			for _, bookmarkID := range bookmarkIDs {
				for _, tagID := range tagIDs {
					_, err := stmtInsertBookTag.ExecContext(ctx, bookmarkID, tagID)
					if err != nil {
						return fmt.Errorf("failed to insert bookmark tag: %w", err)
					}
				}
			}
		}

		return nil
	})
}
