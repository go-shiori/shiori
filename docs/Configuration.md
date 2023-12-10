# Configuration

<!-- TOC -->

- [Data Directory](#data-directory)
- [Webroot](#webroot)
  - [Nginx](#nginx)
- [Database](#database)
  - [MySQL](#mysql)
  - [PostgreSQL](#postgresql)

<!-- /TOC -->

## Data Directory

Shiori is designed to work out of the box, but you can change where it stores your bookmarks if you need to.

By default, Shiori saves your bookmarks in one of the following directories:

| Platform | Directory                                                    |
| -------- | ------------------------------------------------------------ |
| Linux    | `${XDG_DATA_HOME}/shiori` (default: `~/.local/share/shiori`) |
| macOS    | `~/Library/Application Support/shiori`                       |
| Windows  | `%LOCALAPPDATA%/shiori`                                      |

If you pass the flag `--portable` to Shiori, your data will be stored  in the `shiori-data` subdirectory alongside the shiori executable.

To specify a custom path, set the `SHIORI_DIR` environment variable.

## Webroot

If you want to serve Shiori behind a reverse proxy, you can set the `SHIORI_WEBROOT` environment variable to the path where Shiori is served, e.g. `/shiori`.

Keep in mind this configuration wont make Shiori accessible from `/shiori` path so you need to setup your reverse proxy accordingly so it can strip the webroot path.

We provide some examples for popular reverse proxies below. Please follow your reverse proxy documentation in order to setup it properly.

### Nginx

Fox nginx, you can use the following configuration as a example. The important part **is the trailing slash in `proxy_pass` directive**:

```nginx
location /shiori {
    proxy_pass http://localhost:8080/;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
}
```

## Database

Shiori uses an SQLite3 database stored in the above data directory by default. If you prefer, you can also use MySQL or PostgreSQL database by setting it in environment variables.

### MySQL

MySQL example: `SHIORI_DATABASE_URL="mysql://username:password@(hostname:port)/database?charset=utf8mb4"`

You can find additional details in [go mysql sql driver documentation](https://github.com/go-sql-driver/mysql#dsn-data-source-name).

### PostgreSQL

PostgreSQL example: `SHIORI_DATABASE_URL="postgres://pqgotest:password@hostname/database?sslmode=verify-full"`

You can find additional details in [go postgres sql driver documentation](https://pkg.go.dev/github.com/lib/pq).
