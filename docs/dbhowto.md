# Shirori Database Howto (Non sqlite)

By default Shirori keeps all of its bookmark data in a sqlite database hosted on the filesystem. If you want to use an external database, this document explains how to use either MYSQL or Postgres.

## Assumptions

1. That you have properly configured your database server. We won't tell you how do this.
2. That you have properly secured your database. We won't tell you how to do this
3. That you know the relevant database connection settings, database names etc. and have already taken care of this We won't tell you how to do this
4. That you are using either postgres or mysql

## Postgresql

Shiori reads environment variables on startup. You will need to have the following environment variables set prior to starting shiori:

```bash
	export SHIORI_DBMS = "postgresql"
	export SHIORI_PG_HOST = "your.host.here"
	export SHIORI_PG_PORT = "your port here"
	export SHIORI_PG_USER = "yourpostgresuser"
	export SHIORI_PG_PASS = "yoursupersecretpassword"
	export SHIORI_PG_NAME = "your_database_name"
```

## Mysql

Shiori reads environment variables on startup. You will need to have the following environment variables configured prior to starting shiori:

```bash
export SHIORI_DBMS = "mysql"
export SHIORI_MYSQL_USER = "youruser"
export SHIORI_MYSQL_PASS "yoursecretpassword"
export SHIORI_MYSQL_NAME = "yourdatabasename"
export SHIORI_MYSQL_ADDRESS = "my.sql.server.address"
```

## Disclaimers and warnings

To re-iterate, we do not support teaching you how to use databases, how to secure them, or how to set environment variables for things like zsh, or fish. We assume 'bash' with relatively common databases etc. You will need to backup databases etc. on remote, since you will no longer have a local copy. Things like database secrets shoußld be securely injected with appropriate controls etc. ***YOU HAVE BEEN WARNED.***
