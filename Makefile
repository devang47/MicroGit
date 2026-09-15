.PHONY: build test race lint fmt vet cover install clean

# Build the microgit binary into ./microgit.
build:
	go build -o microgit .

# Run the full test suite (unit + e2e).
test:
	go test ./...

# Run tests with the race detector.
race:
	go test -race ./...

# Report test coverage per package.
cover:
	go test -cover ./...

# Format, vet and (if installed) run staticcheck.
lint: fmt vet
	@command -v staticcheck >/dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed, skipping"

fmt:
	gofmt -l -w .

vet:
	go vet ./...

# Install microgit into $GOBIN / $GOPATH/bin.
install:
	go install .

clean:
	rm -f microgit
	rm -rf dist
