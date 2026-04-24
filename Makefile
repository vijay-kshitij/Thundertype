BINARY=thundertype
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

.PHONY: build run clean

# Build the binary
build:
	CGO_ENABLED=1 go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY) .

# Build and run
run: build
	./$(BINARY)

# Quick run (compile + run in one step, good for development)
dev:
	go run .

# Clean build artifacts
clean:
	rm -f $(BINARY)

# Build for both Intel and Apple Silicon
universal:
	CGO_ENABLED=1 GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY)-arm64 .
	CGO_ENABLED=1 GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY)-amd64 .
	lipo -create -output $(BINARY) $(BINARY)-arm64 $(BINARY)-amd64
	rm $(BINARY)-arm64 $(BINARY)-amd64
