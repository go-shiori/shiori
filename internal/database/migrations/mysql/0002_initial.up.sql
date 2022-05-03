CREATE TABLE IF NOT EXISTS bookmark(
		id       INT(11)    NOT NULL AUTO_INCREMENT,
		url      TEXT       NOT NULL,
		title    TEXT       NOT NULL,
		excerpt  TEXT       NOT NULL DEFAULT (''),
		author   TEXT       NOT NULL DEFAULT (''),
		public   BOOLEAN    NOT NULL DEFAULT 0,
		content  MEDIUMTEXT NOT NULL DEFAULT (''),
		html     MEDIUMTEXT NOT NULL DEFAULT (''),
		modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		PRIMARY KEY(id),
		UNIQUE KEY bookmark_url_UNIQUE (url(255)),
		FULLTEXT (title, excerpt, content))
		CHARACTER SET utf8mb4;