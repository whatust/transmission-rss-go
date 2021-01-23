LOG_DIR = "test/log"

build:
	go build -o bin/main main.go

run:
	go run main.go --config local_test/config/config.yml

.PHONY: clean
clean:
	rm -rf ${LOG_DIR}/transmission-rss.log ${LOG_DIR}/transmission-rss-seen.log

.PHONY: test
test:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...