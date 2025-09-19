# This is the Go package to run a target against. Useful for running tests of one package, for example.
PACKAGE ?= ./...

# We want to have our binaries in the bin subdirectory available. In addition we want them to have priority over
# binaries somewhere else on the system.
export PATH := $(CURDIR)/bin:$(PATH)

.PHONY: all
all: test

.PHONY: test
test: prepare
	go test $(PACKAGE)

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
