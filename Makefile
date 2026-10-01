.PHONY: build install test vet check-specs release clean

# Pure Go: a static binary that needs no libc (specs/001-releases)
export CGO_ENABLED := 0

BINARY := i3qws

# Version from git: the last tag, commits since it, -dirty; "dev" outside a
# repository (specs/001-releases)
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/cured-plumbum/i3qws/cmd.version=$(VERSION)

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) .

# The same build into $(go env GOPATH)/bin, or GOBIN when set
install:
	go install -trimpath -ldflags "$(LDFLAGS)" .

test:
	go test ./...

vet:
	go vet ./...

# The form of specs/
check-specs:
	scripts/check-specs.sh

# GitHub release on a pushed version tag: make release TAG=vX.Y.Z
release:
	scripts/release.sh $(TAG)

clean:
	rm -f $(BINARY)
