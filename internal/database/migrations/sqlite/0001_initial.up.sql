CREATE TABLE IF NOT EXISTS account(
    id INTEGER NOT NULL,
    username TEXT NOT NULL,
    password TEXT NOT NULL,
    owner INTEGER NOT NULL DEFAULT 0,
    CONSTRAINT account_PK PRIMARY KEY(id),
    CONSTRAINT account_username_UNIQUE UNIQUE(username)
);

CREATE TABLE IF NOT EXISTS bookmark(
    id INTEGER NOT NULL,
    url TEXT NOT NULL,
    title TEXT NOT NULL,
    excerpt TEXT NOT NULL DEFAULT "",
    author TEXT NOT NULL DEFAULT "",
    public INTEGER NOT NULL DEFAULT 0,
    modified TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT bookmark_PK PRIMARY KEY(id),
    CONSTRAINT bookmark_url_UNIQUE UNIQUE(url)
);

CREATE TABLE IF NOT EXISTS tag(
    id INTEGER NOT NULL,
    name TEXT NOT NULL,
    CONSTRAINT tag_PK PRIMARY KEY(id),
    CONSTRAINT tag_name_UNIQUE UNIQUE(name)
);

CREATE TABLE IF NOT EXISTS bookmark_tag(
    bookmark_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    CONSTRAINT bookmark_tag_PK PRIMARY KEY(bookmark_id, tag_id),
    CONSTRAINT bookmark_id_FK FOREIGN KEY(bookmark_id) REFERENCES bookmark(id),
    CONSTRAINT tag_id_FK FOREIGN KEY(tag_id) REFERENCES tag(id)
);

CREATE VIRTUAL TABLE IF NOT EXISTS bookmark_content
    USING fts5(title, content, html, docid);
