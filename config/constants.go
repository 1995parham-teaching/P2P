package config

import "time"

// Buffer sizes for network operations
const (
	// BufferSize is the standard buffer size for file transfer operations
	BufferSize = 1024

	// UDPBufferSize is the buffer size for UDP message reading
	UDPBufferSize = 2048

	// FileNameLength is the fixed length for file name in protocol
	FileNameLength = 64

	// FileSizeLength is the fixed length for file size in protocol
	FileSizeLength = 10
)

// Protocol constants
const (
	// TransferMethodTCP indicates TCP-based file transfer
	TransferMethodTCP = 1

	// TransferMethodReliableUDP indicates reliable UDP-based file transfer
	TransferMethodReliableUDP = 2
)

// Timing constants
const (
	// NonPriorResponseDelay is the delay for non-priority responders
	NonPriorResponseDelay = 10 * time.Second
)

// Message type constants
const (
	MsgDiscover = "DISCOVER"
	MsgGet      = "Get"
	MsgFile     = "File"
	MsgStopWait = "SW"
	MsgAsk      = "Ask"
	MsgFileSize = "FSize"
	MsgFileName = "Name"
	MsgBuffer   = "buffer"
	MsgAck      = "Ack"
)
