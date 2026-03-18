BINARY_NAME := enter
VERSION := $(shell cat VERSION)
LDFLAGS := "-X main.version=$(VERSION) -X main.revision=$(shell git rev-parse --verify --short HEAD)"

all: build

test:
	go test ./...

build:
	go build -ldflags $(LDFLAGS) -trimpath -o $(BINARY_NAME) ./cmd/enter/

install:
	go install -ldflags $(LDFLAGS) ./cmd/enter/

clean:
	rm -f $(BINARY_NAME)

.PHONY: all test build install clean
