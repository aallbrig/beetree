VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.1.0")
GOFLAGS := -ldflags "-X github.com/aallbrig/beetree-cli/cmd.Version=$(VERSION)"
GOBIN ?= $(shell go env GOPATH)/bin

.PHONY: build install test clean demo demo-clean

build:
	cd cli && go build $(GOFLAGS) -o ../bin/beetree .

install: build
	cp bin/beetree $(GOBIN)/beetree

test:
	cd cli && go test -count=1 ./...

clean:
	rm -rf bin/ cli/generated/

demo: install
	./scripts/record-demo.sh

demo-clean:
	rm -rf demos/segments/*.mp4 dist/demo.mp4
