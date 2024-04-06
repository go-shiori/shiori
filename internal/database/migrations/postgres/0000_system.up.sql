CREATE TABLE IF NOT EXISTS shiori_system(
    database_version TEXT NOT NULL DEFAULT '0.0.0'
);

INSERT INTO shiori_system(database_version) VALUES('0.0.0');
