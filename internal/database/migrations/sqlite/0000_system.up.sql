CREATE TABLE IF NOT EXISTS shiori_system(
    database_schema_version TEXT NOT NULL DEFAULT '0.0.0'
);

INSERT INTO shiori_system(database_schema_version) VALUES('0.0.0');
