ALTER TABLE bookmark
ADD COLUMN created TEXT;


UPDATE bookmark
SET created = bookmark.modified
WHERE modified IS NOT NULL;
