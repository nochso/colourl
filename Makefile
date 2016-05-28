# These are the values we want to pass for Version and BuildTime
VERSION := $(shell git describe --tags --always --dirty)
BUILD_DATE := $(shell date -u "+%F %T %Z")

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags "-X 'main.Version=${VERSION}' -X 'main.BuildDate=${BUILD_DATE}'"

build:
	go generate ./...
	go build ${LDFLAGS} ./cmd/colourl-http

test:
	go test ./...