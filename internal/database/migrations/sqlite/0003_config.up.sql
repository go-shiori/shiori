ALTER TABLE account
    ADD config TEXT NOT NULL DEFAULT '{"showId":false,"listMode":false,"hideThumbnail":false,"hideExcerpt":false,"nightMode":false,"keepMetadata":false,"useArchive":false,"makePublic":false}';

