BIN := bento
PREFIX ?= $(HOME)/.local
VERSION ?= 0.1.0
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build install test darwin-arm64 darwin-amd64 clean

build:
	go build -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/bento

install: build
	mkdir -p $(PREFIX)/bin
	cp $(BIN) $(PREFIX)/bin/$(BIN)

test:
	go test ./...

darwin-arm64:
	mkdir -p dist
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/bento_darwin_arm64 ./cmd/bento

darwin-amd64:
	mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/bento_darwin_amd64 ./cmd/bento

clean:
	rm -f $(BIN)
	rm -rf dist
