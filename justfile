# P2P File Sharing - Just Commands
# https://github.com/casey/just

# Default recipe to display help
default:
    @just --list

# Binary name

binary := "p2p"

# Build the application
build:
    go build -o {{ binary }} ./cmd/p2p

# Run the application
run: build
    ./{{ binary }}

# Run tests
test:
    go test -v ./...

# Run tests with coverage
test-coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
    rm -f {{ binary }} coverage.out coverage.html
    go clean

# Format code
fmt:
    go fmt ./...

# Run linter
lint:
    golangci-lint run

# Build Docker image
docker-build:
    docker compose build

# Start all nodes with Docker Compose
docker-up:
    docker compose up -d

# Stop all nodes
docker-down:
    docker compose down

# View logs from all nodes
docker-logs:
    docker compose logs -f

# Attach to a specific node (usage: just attach node1)
attach node="node1":
    docker attach p2p-{{ node }}

# Setup demo files
demo-setup:
    mkdir -p demo/node1 demo/node2 demo/node3
    echo "Hello from Node 1!" > demo/node1/hello-from-node1.txt
    echo "Hello from Node 2!" > demo/node2/hello-from-node2.txt
    echo "Hello from Node 3!" > demo/node3/hello-from-node3.txt
    @echo "Demo files created in demo/ directory"

# Full demo: build and start
demo: docker-build docker-up
    @echo ""
    @echo "=== P2P Demo Started ==="
    @echo ""
    @echo "Three nodes are running:"
    @echo "  - node1 (sharing: demo/node1/)"
    @echo "  - node2 (sharing: demo/node2/)"
    @echo "  - node3 (sharing: demo/node3/)"
    @echo ""
    @echo "To interact with a node:"
    @echo "  just attach node1"
    @echo "  just attach node2"
    @echo "  just attach node3"
    @echo ""
    @echo "Available commands inside a node:"
    @echo "  list              - Show cluster members"
    @echo "  get <filename>    - Download a file"
    @echo "  quit              - Exit the node"
    @echo ""
    @echo "Example: From node1, run 'get hello-from-node2.txt'"
    @echo ""
    @echo "To stop the demo: just docker-down"

# Stop demo (alias)
demo-stop: docker-down
