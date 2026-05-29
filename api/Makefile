BINARY=apc
BUILD_DIR=./bin

.PHONY: build test vet clean

build:
	go build -o $(BUILD_DIR)/$(BINARY) .

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)
