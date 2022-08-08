BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS bookmark_temp(
    id INTEGER NOT NULL,
    url TEXT NOT NULL,
    title TEXT NOT NULL,
    excerpt TEXT NOT NULL DEFAULT '',
    author TEXT NOT NULL DEFAULT '',
    public INTEGER NOT NULL DEFAULT 0,
    modified TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT bookmark_PK PRIMARY KEY(id),
    CONSTRAINT bookmark_url_UNIQUE UNIQUE(url)
);
INSERT INTO bookmark_temp SELECT id, url, title, excerpt, author, public, modified FROM bookmark;
DROP TABLE bookmark;
ALTER TABLE bookmark_temp RENAME TO bookmark;
COMMIT;
