-- Create a temporary table
CREATE TABLE IF NOT EXISTS bookmark_temp(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url TEXT NOT NULL,
    title TEXT NOT NULL,
    excerpt TEXT NOT NULL DEFAULT "",
    author TEXT NOT NULL DEFAULT "",
    public INTEGER NOT NULL DEFAULT 0,
    modified TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    has_content BOOLEAN DEFAULT FALSE NOT NULL,
    CONSTRAINT bookmark_url_UNIQUE UNIQUE(url)
);

-- Copy data from the original table to the temporary table
INSERT INTO bookmark_temp (id, url, title, excerpt, author, public, modified, has_content)
SELECT id, url, title, excerpt, author, public, modified, has_content FROM bookmark;

-- Drop the original table
DROP TABLE bookmark;

-- Rename the temporary table to the original table name
ALTER TABLE bookmark_temp RENAME TO bookmark;
