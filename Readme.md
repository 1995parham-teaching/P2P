# P2P File Sharing

A peer-to-peer file sharing application demonstrating socket programming concepts in Go. Nodes discover each other, share cluster membership, and transfer files using either TCP or a reliable UDP protocol.

## Architecture Overview

```text
┌─────────────────────────────────────────────────────────────┐
│                          Node                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  UDP Server │  │  TCP Server │  │  Reliable UDP Server│  │
│  │  (Discovery │  │  (File      │  │  (Stop-and-Wait     │  │
│  │   & Control)│  │   Transfer) │  │   File Transfer)    │  │
│  └──────┬──────┘  └──────┬──────┘  └──────────┬──────────┘  │
│         │                │                    │             │
│         └────────────────┼────────────────────┘             │
│                          │                                  │
│                    ┌─────┴─────┐                            │
│                    │  Cluster  │                            │
│                    │  Manager  │                            │
│                    └───────────┘                            │
└─────────────────────────────────────────────────────────────┘
```

Each node runs three servers:

- **UDP Server**: Handles peer discovery and file request coordination
- **TCP Server**: Serves files to requesting peers (reliable, built-in)
- **Reliable UDP Server**: Alternative file transfer using stop-and-wait protocol

## How It Works

### 1. Cluster Discovery

Nodes maintain a list of known peers (the "cluster"). Periodically, each node broadcasts its cluster list to all known peers via UDP:

```
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

```
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
    │ 3. File,1,33680            │                             │
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
3. Nodes that have the file respond with a `File` message containing the transfer method and port
4. Node A connects to the first responder
5. File is transferred

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
| File     | `File,method,port`                 | Respond that file is available  |

**Method values:**

- `1` = TCP transfer
- `2` = Reliable UDP transfer

### TCP File Transfer Protocol

```
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

### Reliable UDP Messages (Stop-and-Wait)

| Message  | Format               | Description                         |
| -------- | -------------------- | ----------------------------------- |
| Size     | `Size,bytes`         | File size in bytes                  |
| FileName | `Name,filename`      | Name of the file                    |
| Segment  | `Segment,base64data` | File chunk (base64 encoded)         |
| Ack      | `Ack,seq`            | Acknowledgment with sequence number |

## Configuration

Configuration can be set via `config.yml` or environment variables (prefixed with `P2P_`):

```yaml
host: "127.0.0.1" # Node's IP address
port: 1378 # UDP port for discovery
period: 20 # Discovery broadcast interval (seconds)
waiting: 100 # File request timeout (seconds)
type: 1 # Transfer method: 1=TCP, 2=Reliable UDP
addr: "127.0.0.1:1999" # Reliable UDP server address
```

## Project Structure

```
P2P/
├── main.go                 # Entry point
├── cmd/
│   ├── root.go             # CLI setup (Cobra)
│   └── client/client.go    # Node command handler
├── node/node.go            # Main orchestration
├── config/
│   ├── config.go           # Configuration loading (Viper)
│   ├── constants.go        # Shared constants
│   └── default.go          # Default config values
├── cluster/cluster.go      # Thread-safe peer list management
├── message/message.go      # Protocol message types and parsing
├── internal/utils/utils.go # Helper functions
├── tcp/
│   ├── server/server.go    # TCP file server
│   └── client/client.go    # TCP file client
└── udp/
    └── server/server.go    # UDP discovery and coordination
```

## Running a Node

```bash
# Build
go build -o p2p

# Run
./p2p node
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
$ ./p2p node
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
$ ./p2p node
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

## Reliable Data Transfer Methods

### 1. TCP (Method 1)

Uses Go's built-in TCP for reliable, ordered delivery. The OS handles:

- Connection establishment (3-way handshake)
- Flow control
- Congestion control
- Retransmission of lost packets

### 2. Stop-and-Wait over UDP (Method 2)

Implements reliability on top of UDP using alternating bit protocol:

```
Sender                                    Receiver
   │                                          │
   │──────Segment(seq=0, data)───────────────>│
   │                                          │
   │<─────────────Ack(seq=0)──────────────────│
   │                                          │
   │──────Segment(seq=1, data)───────────────>│
   │                                          │
   │<─────────────Ack(seq=1)──────────────────│
   │                                          │
   │──────Segment(seq=0, data)───────────────>│
   │              (lost)                      │
   │                                          │
   │  (timeout, retransmit)                   │
   │──────Segment(seq=0, data)───────────────>│
   │                                          │
   │<─────────────Ack(seq=0)──────────────────│
```

The sequence number alternates between 0 and 1. If an ACK isn't received within the timeout, the segment is retransmitted.

## Security Considerations

- **Path Traversal Protection**: All file paths are sanitized using `filepath.Base()` to prevent `../` attacks
- **File Index**: Files are indexed by name only; subdirectory structure is flattened for sharing
- **No Authentication**: This is a learning project; production use would require authentication and encryption
