# Configuration

<!-- TOC -->

- [Overall Configuration](#overall-configuration)
  - [Global configuration](#global-configuration)
  - [HTTP configuration variables](#http-configuration-variables)
  - [Storage Configuration](#storage-configuration)
    - [The data Directory](#the-data-directory)
  - [Database Configuration](#database-configuration)
    - [MySQL](#mysql)
    - [PostgreSQL](#postgresql)
- [Reverse proxies and the webroot path](#reverse-proxies-and-the-webroot-path)
  - [Nginx](#nginx)

<!-- /TOC -->

## Overall Configuration

Most configuration can be set directly using environment variables or flags. The available flags can be found by running `shiori --help`. The available environment variables are listed below.

### Global configuration

| Environment variable | Default | Required | Description                            |
| -------------------- | ------- | -------- | -------------------------------------- |
| `SHIORI_DEVELOPMENT` | `False` | No       | Specifies if the server is in dev mode |

### HTTP configuration variables

| Environment variable                       | Default | Required | Description                                           |
| ------------------------------------------ | ------- | -------- | ----------------------------------------------------- |
| `SHIORI_HTTP_ENABLED`                      | True    | No       | Enable HTTP service                                   |
| `SHIORI_HTTP_PORT`                         | 8080    | No       | Port number for the HTTP service                      |
| `SHIORI_HTTP_ADDRESS`                      | :       | No       | Address for the HTTP service                          |
| `SHIORI_HTTP_ROOT_PATH`                    | /       | No       | Root path for the HTTP service                        |
| `SHIORI_HTTP_ACCESS_LOG`                   | True    | No       | Logging accessibility for HTTP requests               |
| `SHIORI_HTTP_SERVE_WEB_UI`                 | True    | No       | Serving Web UI via HTTP. Disable serves only the API. |
| `SHIORI_HTTP_SECRET_KEY`                   |         | **Yes**  | Secret key for HTTP sessions.                         |
| `SHIORI_HTTP_BODY_LIMIT`                   | 1024    | No       | Limit for request body size                           |
| `SHIORI_HTTP_READ_TIMEOUT`                 | 10s     | No       | Maximum duration for reading the entire request       |
| `SHIORI_HTTP_WRITE_TIMEOUT`                | 10s     | No       | Maximum duration before timing out writes             |
| `SHIORI_HTTP_IDLE_TIMEOUT`                 | 10s     | No       | Maximum amount of time to wait for the next request   |
| `SHIORI_HTTP_DISABLE_KEEP_ALIVE`           | true    | No       | Disable HTTP keep-alive connections                   |
| `SHIORI_HTTP_DISABLE_PARSE_MULTIPART_FORM` | true    | No       | Disable pre-parsing of multipart form                 |

### Storage Configuration

The `StorageConfig` struct contains settings related to storage.

| Environment variable | Default       | Required | Description                             |
| -------------------- | ------------- | -------- | --------------------------------------- |
| `SHIORI_DIR`         | (current dir) | No       | Directory where Shiori stores its data. |

#### The data Directory

Shiori is designed to work out of the box, but you can change where it stores your bookmarks if you need to.

By default, Shiori saves your bookmarks in one of the following directories:

| Platform | Directory                                                    |
| -------- | ------------------------------------------------------------ |
| Linux    | `${XDG_DATA_HOME}/shiori` (default: `~/.local/share/shiori`) |
| macOS    | `~/Library/Application Support/shiori`                       |
| Windows  | `%LOCALAPPDATA%/shiori`                                      |

If you pass the flag `--portable` to Shiori, your data will be stored  in the `shiori-data` subdirectory alongside the shiori executable.

To specify a custom path, set the `SHIORI_DIR` environment variable.

### Database Configuration

| Environment variable       | Default | Required | Description                                     |
| -------------------------- | ------- | -------- | ----------------------------------------------- |
| `SHIORI_DBMS` (deprecated) | `DBMS`  | No       | Deprecated (Use environment variables for DBMS) |
| `SHIORI_DATABASE_URL`      | `URL`   | No       | URL for the database (required)                 |

> `SHIORI_DBMS` is deprecated and will be removed in a future release. Please use `SHIORI_DATABASE_URL` instead.

Shiori uses an SQLite3 database stored in the above [data directory by default](#storage-configuration). If you prefer, you can also use MySQL or PostgreSQL database by setting the `SHIORI_DATABASE_URL` environment variable.

#### MySQL

MySQL example: `SHIORI_DATABASE_URL="mysql://username:password@(hostname:port)/database?charset=utf8mb4"`

You can find additional details in [go mysql sql driver documentation](https://github.com/go-sql-driver/mysql#dsn-data-source-name).

#### PostgreSQL

PostgreSQL example: `SHIORI_DATABASE_URL="postgres://pqgotest:password@hostname/database?sslmode=verify-full"`

You can find additional details in [go postgres sql driver documentation](https://pkg.go.dev/github.com/lib/pq).

## Reverse proxies and the webroot path

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
