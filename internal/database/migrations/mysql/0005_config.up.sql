ALTER TABLE account
    ADD COLUMN config VARCHAR(500) NOT NULL DEFAULT '{"showId":false,"listMode":false,"hideThumbnail":false,"hideExcerpt":false,"nightMode":false,"keepMetadata":false,"useArchive":false,"makePublic":false}';
