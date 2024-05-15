-- Rename "modified" column to "created_at"
ALTER TABLE bookmark
RENAME COLUMN modified to created_at;

-- Add the "modified_at" column to the bookmark table
ALTER TABLE bookmark
ADD COLUMN modified_at TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- Update the "modified_at" column with the value from the "created_at" column if it is not null
UPDATE bookmark
SET modified_at = COALESCE(created_at, CURRENT_TIMESTAMP)
WHERE created_at IS NOT NULL;

-- Index for "created_at" "modified_at""
CREATE INDEX idx_created_at ON bookmark(created_at);
CREATE INDEX idx_modified_at ON bookmark(modified_at);
