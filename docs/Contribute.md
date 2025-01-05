# Contribute

1. [Running the server locally](#running-the-server-locally)
2. [Updating the API documentation](#updating-the-api-documentation)
3. [Lint the code](#lint-the-code)
4. [Running tests](#running-tests)

## Running the server locally

To run the current development server with the defaults you can run the following command:

```bash
make run-server
```

## Updating the API documentation

> **ℹ️ Note:** This only applies for the Rest API documentation under the `internal/http` folder, **not** the one under `internal/webserver`.

If you make any changes to the Rest API endpoints, you need to update the swagger documentation. In order to do that, you need to have installed [swag](https://github.com/swaggo/swag).

Then, run the following command:

```bash
make swagger
```

## Updating the frontend styles

The styles that are bundled with Shiori are stored under `internal/view/assets/css/style.css` and `internal/view/assets/css/archive.css` and created from the less files under `internal/views/assets/less`.

If you want to make frontend changes you need to do that under the less files and then compile them to css. In order to do that, you need to have installed [bun](https://bun.sh).

Then, run the following command:

```bash
make styles
```

The `style.css`/`archive.css` will be updated and changes **needs to be committed** to the repository.

## Lint the code

In order to lint the code, you need to have installed [golangci-lint](https://golangci-lint.run) and [swag](https://github.com/swaggo/swag).

After that, run the following command:

```bash
make lint
```

If any errors are found please fix them before submitting your PR.

## Running tests

In order to run the test suite, you need to have running a local instance of MariaDB and PostgreSQL.
If you have docker, you can do this by running the following command with the compose file provided:

```bash
docker-compose up -d mariadb mysql postgres
```

After that, provide the environment variables for the unitest to connect to the database engines:

- `SHIORI_TEST_MYSQL_URL` for MySQL
- `SHIORI_TEST_MARIADB_URL` for MariaDB
- `SHIORI_TEST_PG_URL` for PostgreSQL

```
SHIORI_TEST_PG_URL=postgres://shiori:shiori@127.0.0.1:5432/shiori?sslmode=disable
SHIORI_TEST_MYSQL_URL=shiori:shiori@tcp(127.0.0.1:3306)/shiori
SHIORI_TEST_MARIADB_URL=shiori:shiori@tcp(127.0.0.1:3307)/shiori
```

Finally, run the tests with the following command:

```bash
make unittest
```

## Building the documentation

The documentation is built using MkDocs with the Material theme. For installation instructions, please refer to the [MkDocs installation guide](https://www.mkdocs.org/user-guide/installation/).

To preview the documentation locally while making changes, run:

```bash
mkdocs serve
```

This will start a local server at `http://127.0.0.1:8000` where you can preview your changes in real-time.

Documentation for production is generated automatically on every release and published using github pages.

## Running the server with docker

To run the development server using Docker, you can use the provided `docker-compose.yaml` file which includes both PostgreSQL and MariaDB databases:

```bash
docker compose up shiori
```

This will start the Shiori server on port 8080 with hot-reload enabled. Any changes you make to the code will automatically rebuild and restart the server.

By default, it uses SQLite mounting the local `dev-data` folder in the source code path. To use MariaDB or PostgreSQL instead, uncomment the `SHIORI_DATABASE_URL` line for the appropriate engine in the `docker-compose.yaml` file.

## Running the server using an nginx reverse proxy and a custom webroot

To test Shiori behind an nginx reverse proxy with a custom webroot (e.g., `/shiori/`), you can use the provided nginx configuration:

1. First, ensure the `SHIORI_HTTP_ROOT_PATH` environment variable is uncommented in `docker-compose.yaml`:
   ```yaml
   SHIORI_HTTP_ROOT_PATH: /shiori/
   ```

2. Then start both Shiori and nginx services:
   ```bash
   docker compose up shiori nginx
   ```

This will start the shiori service along with nginx. You can access Shiori using [http://localhost:8081/shiori](http://localhost:8081/shiori).

The nginx configuration in `testdata/nginx.conf` handles all the necessary configuration.
