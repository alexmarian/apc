BINARY=apc
BUILD_DIR=./bin

.PHONY: build test vet clean

build:
	CGO_ENABLED=1 go build -o $(BUILD_DIR)/$(BINARY) .

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)
