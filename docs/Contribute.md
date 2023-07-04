# Contribute

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
