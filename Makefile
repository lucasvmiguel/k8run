VERSION ?= 0.0.5

test:
	go test -v ./...

build:
	go build -o bin/ ./...

release:
	git tag -a v$(VERSION) -m "Release version $(VERSION)"
	git push origin v$(VERSION)