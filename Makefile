VERSION ?= 0.0.2

test:
	go test -v ./...

build:
	go build -o bin/ ./...