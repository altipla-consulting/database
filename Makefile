
FILES = $(shell find . -type f -name '*.go' -not -path './vendor/*')

gofmt:
	@gofmt -w $(FILES)
	@gofmt -r '&α{} -> new(α)' -w $(FILES)

deps:
	go get -u github.com/mgechev/revive

	go get -u github.com/go-sql-driver/mysql
	go get -u github.com/stretchr/testify/require

test:
	revive -formatter friendly
	go install .

	docker-compose up -d database
	bash -c "until mysql -h 127.0.0.1 -P 3307 -u dev-user -pdev-password -e ';' 2> /dev/null ; do sleep 1; done"

	go test
