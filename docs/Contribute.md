# Contribute

1. [Running the server locally](#running-the-server-locally)
2. [Updating the API documentation](#updating-the-api-documentation)
3. [Lint the code](#lint-the-code)
4. [Running tests](#running-tests)

## Running the server locally

To run the current development server with the defaults you can run the following command:

```bash
make serve
```

If you want to run the refactored server, you can run the following command:

```bash
make run-server
```

> **ℹ️ Note:** For more information into what the _refactored server_ means, please check this issue: https://github.com/go-shiori/shiori/issues/640

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
docker-compose up -d mariadb postgres
```

After that, provide the `SHIORI_TEST_PG_URL` and `SHIORI_TEST_MYSQL_URL` environment variables with the connection string to the databases:

```
SHIORI_TEST_PG_URL=postgres://shiori:shiori@127.0.0.1:5432/shiori?sslmode=disable
SHIORI_TEST_MYSQL_URL=shiori:shiori@tcp(127.0.0.1:3306)/shiori
```

Finally, run the tests with the following command:

```bash
make unittest
```
