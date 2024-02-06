CREATE TABLE IF NOT EXISTS account(
		id       SERIAL,
		username VARCHAR(250) NOT NULL,
		password BYTEA    NOT NULL,
		owner    BOOLEAN  NOT NULL DEFAULT FALSE,
		PRIMARY KEY (id),
		CONSTRAINT account_username_UNIQUE UNIQUE (username));

CREATE TABLE IF NOT EXISTS bookmark(
		id       SERIAL,
		url      TEXT       NOT NULL,
		title    TEXT       NOT NULL,
		excerpt  TEXT       NOT NULL DEFAULT '',
		author   TEXT       NOT NULL DEFAULT '',
		public   SMALLINT   NOT NULL DEFAULT 0,
		content  TEXT       NOT NULL DEFAULT '',
		html     TEXT       NOT NULL DEFAULT '',
		modified TIMESTAMP(0) NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY(id),
		CONSTRAINT bookmark_url_UNIQUE UNIQUE (url));

CREATE TABLE IF NOT EXISTS tag(
		id   SERIAL,
		name VARCHAR(250) NOT NULL,
		PRIMARY KEY (id),
		CONSTRAINT tag_name_UNIQUE UNIQUE (name));

CREATE TABLE IF NOT EXISTS bookmark_tag(
		bookmark_id INT      NOT NULL,
		tag_id      INT      NOT NULL,
		PRIMARY KEY(bookmark_id, tag_id),
		CONSTRAINT bookmark_tag_bookmark_id_FK FOREIGN KEY (bookmark_id) REFERENCES bookmark (id),
		CONSTRAINT bookmark_tag_tag_id_FK FOREIGN KEY (tag_id) REFERENCES tag (id));

CREATE INDEX IF NOT EXISTS bookmark_tag_bookmark_id_FK ON bookmark_tag (bookmark_id);
CREATE INDEX IF NOT EXISTS bookmark_tag_tag_id_FK ON bookmark_tag (tag_id);
