CREATE TABLE IF NOT EXISTS account(
		id       INT(11)        NOT NULL AUTO_INCREMENT,
		username VARCHAR(250)   NOT NULL,
		password BINARY(80)     NOT NULL,
		owner    TINYINT(1)     NOT NULL DEFAULT '0',
		PRIMARY KEY (id),
		UNIQUE KEY account_username_UNIQUE (username))
		CHARACTER SET utf8mb4;
