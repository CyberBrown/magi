VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"
BINARY := magi

.PHONY: build build-arm64 test clean deploy

build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY) ./cmd/magi/

build-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY)-arm64 ./cmd/magi/

test:
	go test ./...

clean:
	rm -f $(BINARY) $(BINARY)-arm64

deploy: build-arm64
	scp $(BINARY)-arm64 prometheus:~/magi
	ssh prometheus 'sudo systemctl restart magi'
