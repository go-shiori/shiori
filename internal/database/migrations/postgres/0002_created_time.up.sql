-- Add the "created" column to the bookmark table
ALTER TABLE bookmark
ADD COLUMN created TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- Update the "created" column with the value from the "modified" column if it is not null
UPDATE bookmark
SET created = COALESCE(modified, CURRENT_TIMESTAMP)
WHERE modified IS NOT NULL;
