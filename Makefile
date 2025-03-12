VERSION ?= 0.0.9
MAIN_GO = main.go

test:
	go test -v ./...

build:
	go build -o bin/ ./...

update-version:
	@if [ ! -f $(MAIN_GO) ]; then \
		echo "Error: $(MAIN_GO) not found!"; \
		exit 1; \
	fi
	@awk -v new_version="$(VERSION)" '/Version:/ { sub(/"[^"]+"/, "\"" new_version "\"") } { print }' $(MAIN_GO) > $(MAIN_GO).tmp && mv $(MAIN_GO).tmp $(MAIN_GO)
	@if grep -q 'Version:' $(MAIN_GO); then \
		echo "✅ Updated version to $(VERSION) in $(MAIN_GO)"; \
	else \
		echo "❌ Failed to update version. Ensure 'Version:' is defined correctly in $(MAIN_GO)"; \
		exit 1; \
	fi
	git commit -am "Update version to $(VERSION)"
	git push origin main

release: update-version
	git tag -a v$(VERSION) -m "Release version $(VERSION)"
	git push origin v$(VERSION)