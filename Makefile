LOG_DIR = "test/log"

build:
	go build -o bin/main main.go

run:
	go run main.go --config test/config/config.yml

clean:
	rm -rf ${LOG_DIR}/transmission-rss.log ${LOG_DIR}/transmission-rss-seen.log