.PHONY: build
build:
	go build -o bin/s3manager

.PHONY: run
run:
	go run

.PHONY: lint
lint:
	golangci-lint run
