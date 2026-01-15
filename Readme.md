# P2P File Sharing

A peer-to-peer file sharing application demonstrating socket programming concepts in Go. Nodes discover each other, share cluster membership, and transfer files using TCP.

## Architecture Overview

```text
┌─────────────────────────────────────────────────────────────┐
│                          Node                               │
│  ┌─────────────┐  ┌─────────────┐                           │
│  │  UDP Server │  │  TCP Server │                           │
│  │  (Discovery │  │  (File      │                           │
│  │   & Control)│  │   Transfer) │                           │
│  └──────┬──────┘  └──────┬──────┘                           │
│         │                │                                  │
│         └────────────────┘                                  │
│                │                                            │
│          ┌─────┴─────┐                                      │
│          │  Cluster  │                                      │
│          │  Manager  │                                      │
│          └───────────┘                                      │
└─────────────────────────────────────────────────────────────┘
```

Each node runs two servers:

- **UDP Server**: Handles peer discovery and file request coordination
- **TCP Server**: Serves files to requesting peers (reliable, built-in)

## How It Works

### 1. Cluster Discovery

Nodes maintain a list of known peers (the "cluster"). Periodically, each node broadcasts its cluster list to all known peers via UDP:

```text
Node A                     Node B                     Node C
   │                          │                          │
   │──DISCOVER,[B,C]─────────>│                          │
   │──DISCOVER,[B,C]──────────────────────────────────-->│
   │                          │                          │
   │<─────────DISCOVER,[A,C]──│                          │
   │                          │──DISCOVER,[A,B]─────────>│
```

When a node receives a `DISCOVER` message, it merges the received list with its own, learning about new peers transitively.

### 2. File Request Flow

When a user wants to download a file:

```text
┌────────┐                    ┌────────┐                    ┌────────┐
│ Node A │                    │ Node B │                    │ Node C │
│(wants  │                    │(has    │                    │(doesn't│
│ file)  │                    │ file)  │                    │ have)  │
└───┬────┘                    └───┬────┘                    └───┬────┘
    │                             │                             │
    │ 1. Broadcast: Get,resume.pdf│                             │
    │────────────────────────────>│                             │
    │─────────────────────────────────────────────────────────->│
    │                             │                             │
    │                             │ 2. Search local files       │
    │                             │    Found!                   │
    │                             │                             │
    │ 3. File,1,33680             │                             │
    │<────────────────────────────│                             │
    │                             │                             │
    │ 4. TCP Connect to B:33680   │                             │
    │────────────────────────────>│                             │
    │                             │                             │
    │ 5. File transfer via TCP    │                             │
    │<════════════════════════════│                             │
```

**Steps:**

1. Node A broadcasts a `Get` message to all cluster members
2. Each node searches its shared folder for the file
3. Nodes that have the file respond with a `File` message containing the TCP port
4. Node A connects to the first responder
5. File is transferred via TCP

### 3. Priority Responders

The system tracks "priority responders" - peers that have responded quickly to previous requests. When responding to a file request:

- **Priority peers**: Respond immediately
- **Non-priority peers**: Wait 10 seconds before responding

This ensures that fast, reliable peers are preferred for file transfers.

## Message Protocol

All messages are newline-terminated strings with comma-separated fields.

### UDP Messages

| Message  | Format                             | Description                     |
| -------- | ---------------------------------- | ------------------------------- |
| Discover | `DISCOVER,ip1:port1,ip2:port2,...` | Share cluster membership        |
| Get      | `Get,filename`                     | Request a file from the cluster |
| File     | `File,1,port`                      | Respond that file is available  |

### TCP File Transfer Protocol

```text
┌──────────────────────────────────────────────────────┐
│ 1. Client sends: Get,filename\n                      │
├──────────────────────────────────────────────────────┤
│ 2. Server sends:                                     │
│    ┌────────────┬────────────────┬─────────────────┐ │
│    │ File Size  │   File Name    │   File Data     │ │
│    │ (10 bytes) │   (64 bytes)   │   (variable)    │ │
│    │ padded ':' │   padded ':'   │                 │ │
│    └────────────┴────────────────┴─────────────────┘ │
└──────────────────────────────────────────────────────┘
```

## Configuration

Configuration can be set via `config.yml` or environment variables (prefixed with `P2P_`):

```yaml
host: "127.0.0.1" # Node's IP address
port: 1378 # UDP port for discovery
period: 20 # Discovery broadcast interval (seconds)
waiting: 100 # File request timeout (seconds)
```

## Project Structure

This project follows the [golang-standards/project-layout](https://github.com/golang-standards/project-layout):

```text
P2P/
├── cmd/
│   └── p2p/
│       └── main.go              # Application entry point
├── configs/
│   └── config.example.yml       # Example configuration file
├── internal/                    # Private application code
│   ├── cluster/
│   │   └── cluster.go           # Thread-safe peer list management
│   ├── config/
│   │   ├── config.go            # Configuration loading (Viper)
│   │   ├── constants.go         # Shared constants
│   │   └── default.go           # Default config values
│   ├── message/
│   │   └── message.go           # Protocol message types and parsing
│   ├── node/
│   │   └── node.go              # Main node orchestration
│   ├── tcp/
│   │   ├── client/
│   │   │   └── client.go        # TCP file download client
│   │   └── server/
│   │       └── server.go        # TCP file server
│   ├── udp/
│   │   └── server/
│   │       └── server.go        # UDP discovery and coordination
│   └── utils/
│       └── utils.go             # Helper functions
├── go.mod
├── go.sum
└── README.md
```

### Directory Descriptions

- **`/cmd`**: Main applications for this project. The directory name for each application should match the name of the executable (e.g., `/cmd/p2p`).

- **`/internal`**: Private application and library code. This is the code you don't want others importing in their applications. Note that this layout pattern is enforced by the Go compiler.

- **`/configs`**: Configuration file templates or default configs. Put your `config.yml` here or in the project root.

## Running a Node

### Local Build

```bash
# Build
go build -o p2p ./cmd/p2p

# Run
./p2p
```

### Using Just

[Just](https://github.com/casey/just) is a command runner. Install it and run:

```bash
just          # Show available commands
just build    # Build the application
just run      # Build and run
just test     # Run tests
```

The node will prompt for:

1. **Shared folder**: Directory containing files to share
2. **Cluster members**: Initial list of peer addresses (IP:port format)

### Commands

Once running, enter commands at the prompt:

| Command          | Description                      |
| ---------------- | -------------------------------- |
| `list`           | Show all known cluster members   |
| `get <filename>` | Download a file from the cluster |
| `quit`           | Shutdown the node gracefully     |

## Example Session

**Terminal 1 (Node A on port 1378):**

```
$ ./p2p
Enter the folder you want to share:
/home/user/shared
Enter your cluster members list (one per line, enter 'q' to finish):
127.0.0.1:1379
q
Enter a file you want to download or 'list' to see the cluster ('quit' to exit)
list
Cluster members:
  1. 127.0.0.1:1379
get resume.pdf
Received file completely!
```

**Terminal 2 (Node B on port 1379):**

```
$ ./p2p
Enter the folder you want to share:
/home/user/documents
Enter your cluster members list (one per line, enter 'q' to finish):
127.0.0.1:1378
q
Enter a file you want to download or 'list' to see the cluster ('quit' to exit)
A client has connected!
Sending filename and filesize!
Start sending file
File has been sent, closing connection!
```

## Docker Demo

The easiest way to demo the P2P application is using Docker Compose, which sets up 3 interconnected nodes automatically.

### Quick Start

```bash
# Start the demo (builds and runs 3 nodes)
just demo

# Or manually:
docker compose up -d
```

### Interacting with Nodes

Each node runs in its own container with a pre-configured shared folder:

```bash
# Attach to node1
just attach node1

# Or directly with docker:
docker attach p2p-node1
```

### Demo Walkthrough

1. **Start the demo:**

   ```bash
   just demo
   ```

2. **Attach to node1:**

   ```bash
   just attach node1
   ```

3. **List cluster members:**

   ```
   list
   Cluster members:
     1. node2:1378
     2. node3:1378
   ```

4. **Download a file from another node:**

   ```
   get hello-from-node2.txt
   Received file completely!
   ```

5. **Check the downloaded file:**

   ```bash
   cat demo/node1/hello-from-node2.txt
   ```

6. **Stop the demo:**
   ```bash
   just docker-down
   ```

### Demo Directory Structure

```
demo/
├── node1/
│   └── hello-from-node1.txt    # Shared by node1
├── node2/
│   └── hello-from-node2.txt    # Shared by node2
└── node3/
    └── hello-from-node3.txt    # Shared by node3
```

### Environment Variables

When running with Docker, the following environment variables configure the node:

| Variable      | Description                            | Example                 |
| ------------- | -------------------------------------- | ----------------------- |
| `P2P_HOST`    | Node's hostname/IP                     | `node1`                 |
| `P2P_PORT`    | UDP port for discovery                 | `1378`                  |
| `P2P_FOLDER`  | Shared folder path                     | `/app/shared`           |
| `P2P_CLUSTER` | Comma-separated list of peer addresses | `node2:1378,node3:1378` |

## Security Considerations

- **Path Traversal Protection**: All file paths are sanitized using `filepath.Base()` to prevent `../` attacks
- **File Index**: Files are indexed by name only; subdirectory structure is flattened for sharing
- **No Authentication**: This is a learning project; production use would require authentication and encryption
