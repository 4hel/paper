.PHONY: test build server client clean end2end

# Default target - run tests with race detection and verbose output
test:
	@echo "Running tests with race detection and verbose output..."
	go test ./internal/... -race -v

# Simple build
build:
	go build -o paperserver ./cmd/paperserver
	go build -o client ./cmd/client

# Run server
server: build
	./paperserver

# Run client (requires name parameter)
client: build
ifndef name
	@echo "Usage: make client name=PlayerName"
	@exit 1
endif
	./client -name "$(name)"

# Run end-to-end tests
end2end:
	@echo "Running end-to-end tests..."
	cd cmd/paperserver && go test -v -timeout=60s

# Clean up
clean:
	rm -f paperserver client