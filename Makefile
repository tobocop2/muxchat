.PHONY: build test clean install

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -ldflags "-X github.com/tobias/muxchat/cmd.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o muxchat .

test:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo ""
	@echo "=== Coverage Report ==="
	@go tool cover -func=coverage.out | tail -1
	@echo ""
	@go tool cover -func=coverage.out | grep -v "total:" | sort -t'%' -k3 -rn | head -10
	@echo "..."
	@echo ""
	@echo "Run 'go tool cover -html=coverage.out' for detailed report"

clean:
	rm -f muxchat

install:
	go install $(LDFLAGS) .

.DEFAULT_GOAL := build
