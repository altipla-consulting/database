
# database

[![GoDoc](https://godoc.org/github.com/altipla-consulting/database?status.svg)](https://godoc.org/github.com/altipla-consulting/database)

> Database helper to read and write models.


### Install

```shell
go get github.com/altipla-consulting/database
```

This library has no external dependencies outside the Go standard library.


### Contributing

You can make pull requests or create issues in GitHub. Any code you send should be formatted using ```gofmt```.


### Running tests

Start the test database:

```shell
docker-compose up -d database
```

Install test libs:

```shell
go get github.com/stretchr/testify
go get github.com/go-sql-driver/mysql
```

Run the tests:

```shell
go test
```

Shutdown the database when finished testing:

```shell
docker-compose stop database
```


### License

[MIT License](LICENSE)
