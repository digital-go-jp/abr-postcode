.PHONY: help build run test clean lint fmt vuln

VERSION ?= $(shell cat VERSION)
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
LDFLAGS := -ldflags "-X abr-postcode/internal/version.Version=$(VERSION) -X abr-postcode/internal/version.Commit=$(COMMIT)"

help:
	@echo "Available targets:"
	@echo "  build  - Build the application"
	@echo "  run    - Run the application"
	@echo "  test   - Run tests"
	@echo "  clean  - Clean build artifacts and downloaded/generated data"
	@echo "  lint   - Run lint checks (no auto-fix; suitable for CI)"
	@echo "  fmt    - Auto-format and fix lint issues"
	@echo "  vuln   - Run govulncheck"

build:
	go build $(LDFLAGS) -o abrp .

run:
	go run $(LDFLAGS) . serve

test:
	go test ./...

clean:
	rm -f abrp
	rm -f data/abr_post_code.zip data/abr_post_code.csv data/metadata.txt \
	      data/city.csv data/town.csv data/post_code_mapping.csv \
	      data/data_modified.txt

lint:
	go mod verify
	go vet ./...
	golangci-lint run

fmt:
	go mod tidy
	goimports -w .
	golangci-lint run --fix

vuln:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
