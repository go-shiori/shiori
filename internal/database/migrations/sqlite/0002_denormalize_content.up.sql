ALTER TABLE bookmark
    ADD has_content BOOLEAN DEFAULT FALSE NOT NULL;

UPDATE bookmark
SET has_content = bc.has_content FROM (SELECT docid, content <> '' AS has_content FROM bookmark_content) AS bc
WHERE bookmark.id = bc.docid;
