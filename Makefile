
FILES = $(shell find . -type f -name '*.go')

gofmt:
	@gofmt -w $(FILES)
	@gofmt -r '&α{} -> new(α)' -w $(FILES)

deps:
	go get -u github.com/mgechev/revive

test: gofmt
	revive -formatter friendly
	go install .

	docker-compose up -d database
	bash -c "until mysql -h 127.0.0.1 -P 3307 -u dev-user -pdev-password -e ';' 2> /dev/null ; do sleep 1; done"

	go test .

update-deps:
	go get -u
	go mod download
	go mod tidy
