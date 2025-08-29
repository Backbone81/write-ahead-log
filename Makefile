.PHONY: all
all: test

.PHONY: test
test:
	go test ./...

.PHONY: benchmark
benchmark:
	go test -bench=. -benchmem ./...
