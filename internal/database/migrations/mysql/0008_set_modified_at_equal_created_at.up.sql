UPDATE bookmark
SET modified_at = COALESCE(created_at, CURRENT_TIMESTAMP)
WHERE created_at IS NOT NULL;
