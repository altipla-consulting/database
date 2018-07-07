
# database

[![GoDoc](https://godoc.org/github.com/altipla-consulting/database?status.svg)](https://godoc.org/github.com/altipla-consulting/database)
[![Build Status](https://travis-ci.org/altipla-consulting/database.svg?branch=master)](https://travis-ci.org/altipla-consulting/database)

Database helper to read and write models.


### Install

```shell
go get github.com/altipla-consulting/database
```

This library has the following dependencies:
- [github.com/go-sql-driver/mysql](github.com/go-sql-driver/mysql)


### Contributing

You can make pull requests or create issues in GitHub. Any code you send should be formatted using ```gofmt```.


### Running tests

Download any dependency your system may need:

```shell
make deps
```

Then run the tests and configure a local MySQL instance for them:

```shell
make test
```


### License

[MIT License](LICENSE)
