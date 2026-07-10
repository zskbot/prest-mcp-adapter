.PHONY: test test-unit vet build fmt coverage

test: test-unit coverage

test-unit:
	go test -timeout 30s -race -count=1 ./...

coverage:
	COVERAGE_MIN=80 bash scripts/check-coverage.sh

vet:
	go vet ./...

build:
	go build -o bin/prest-mcp ./cmd/prest-mcp

fmt:
	gofmt -w .
