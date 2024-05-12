ALTER TABLE bookmark
RENAME COLUMN modified to created_at;

ALTER TABLE bookmark
ADD COLUMN modified_at TEXT NULL;

UPDATE bookmark
SET modified_at = bookmark.created_at
WHERE created_at IS NOT NULL;
