.PHONY: build
build:
	go build -o s3manager ./cmd/s3manager

.PHONY: run
run:
	go run ./cmd/s3manager

.PHONY: lint
lint:
	golangci-lint run

.PHONY: clean-build
clean-build:
	rm -rf ./dist/*