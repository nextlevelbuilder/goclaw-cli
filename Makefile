VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
  -X github.com/nextlevelbuilder/goclaw-cli/cmd.Version=$(VERSION) \
  -X github.com/nextlevelbuilder/goclaw-cli/cmd.Commit=$(COMMIT) \
  -X github.com/nextlevelbuilder/goclaw-cli/cmd.BuildDate=$(DATE)

.PHONY: build test lint install clean

build:
	go build -ldflags "$(LDFLAGS)" -o goclaw .

test:
	go test -race ./...

lint:
	go vet ./...

install:
	go install -ldflags "$(LDFLAGS)" .

clean:
	rm -f goclaw goclaw.exe
	rm -rf dist/
