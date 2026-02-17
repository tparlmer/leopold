.PHONY: test lint build clean

test:
	go test -race -v ./...

lint:
	golangci-lint run ./...

build:
	go build ./...

clean:
	go clean
	rm -f coverage.out

cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
