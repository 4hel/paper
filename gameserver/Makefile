.PHONY: test build server client clean end2end tag

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

# Run client connected to production server (requires name parameter)
client-prd:
ifndef name
	@echo "Usage: make client-prd name=PlayerName"
	@exit 1
endif
	go run ./cmd/client -name="$(name)" -server=paperserver-prd.dingodream.org

# Run end-to-end tests
end2end:
	@echo "Running end-to-end tests..."
	cd cmd/paperserver && go test -v -timeout=60s

# Clean up
clean:
	rm -f paperserver client

git/tag:
	@echo "Creating a new annotated tag..."
	@read -p "Enter tag name (e.g., v1.0.0): " tag_name; \
	read -p "Enter tag message: " tag_message; \
	echo "Creating tag: $$tag_name"; \
	git tag -a "$$tag_name" -m "$$tag_message"; \
	echo "Pushing tag to origin..."; \
	git push origin "$$tag_name"; \
	echo "Tag $$tag_name created and pushed successfully!"