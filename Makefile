# This is the Go package to run a target against. Useful for running tests of one package, for example.
PACKAGE ?= ./...

# We want to have our binaries in the bin subdirectory available. In addition we want them to have priority over
# binaries somewhere else on the system.
export PATH := $(CURDIR)/bin:$(PATH)

.PHONY: all
all: build

.PHONY: build
build: prepare
	go build ./cmd/wal-cli

.PHONY: test
test: prepare
	rm -rf tmp/coverage
	mkdir -p tmp/coverage
	go test --race -coverpkg=./... -cover $(PACKAGE) -args -test.gocoverdir=$(CURDIR)/tmp/coverage
	@echo
	@echo "========== Corrected coverage over all packages =========="
	go tool covdata percent -i=tmp/coverage
	go tool covdata textfmt -i=tmp/coverage -o tmp/cover.out
	go tool cover -html=tmp/cover.out -o tmp/cover.html

.PHONY: benchmark
benchmark: prepare
	go test -run=^$$ -bench=. -benchmem $(PACKAGE)

.PHONY: prepare
prepare:
	go mod tidy
	go fmt $(PACKAGE)
	go vet $(PACKAGE)
	golangci-lint run --fix

.PHONY: clean
clean:
	rm -rf tmp
