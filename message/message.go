package message

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/elahe-dastan/reliable_UDP/message"
	"github.com/1995parham-teaching/P2P/config"
)

var (
	ErrMalformedMessage = errors.New("malformed message")
	ErrUnknownMessage   = errors.New("unknown message type")
	ErrInvalidPort      = errors.New("invalid port number")
	ErrInvalidMethod    = errors.New("invalid transfer method")
	ErrInvalidSize      = errors.New("invalid file size")
)

type Message interface {
	Marshal() string
}

type Discover struct {
	List []string
}

type Get struct {
	Name string
}

type File struct {
	Method  int
	TCPPort int
	UDPPort int
}

type StopWait struct{}

type AskFile struct {
	Name string
}

type Size struct {
	Size int64
}

type FileName struct {
	Name string
}

type Segment struct {
	Part []byte
}

type Acknowledgment struct {
	Seq int
}

func (d *Discover) Marshal() string {
	list := strings.Join(d.List, ",")
	return fmt.Sprintf("%s,%s\n", config.MsgDiscover, list)
}

func (g *Get) Marshal() string {
	return fmt.Sprintf("%s,%s\n", config.MsgGet, g.Name)
}

func (f *File) Marshal() string {
	if f.Method == config.TransferMethodTCP {
		return fmt.Sprintf("%s,%d,%d\n", config.MsgFile, f.Method, f.TCPPort)
	}
	// Reliable UDP
	return fmt.Sprintf("%s,%d,%d\n", config.MsgFile, f.Method, f.UDPPort)
}

func (s *StopWait) Marshal() string {
	return fmt.Sprintf("%s\n", config.MsgStopWait)
}

func (a *AskFile) Marshal() string {
	return fmt.Sprintf("%s,%s\n", config.MsgAsk, a.Name)
}

func (a *Acknowledgment) Marshal() string {
	return fmt.Sprintf("%s,%d\n", config.MsgAck, a.Seq)
}

func (s *Size) Marshal() string {
	return fmt.Sprintf("%s,%d\n", message.Size, s.Size)
}

func (n *FileName) Marshal() string {
	return fmt.Sprintf("%s,%s\n", message.FileName, n.Name)
}

func (s *Segment) Marshal() string {
	return fmt.Sprintf("%s,%s\n", message.Segment, base64.StdEncoding.EncodeToString(s.Part))
}

// Unmarshal parses a message string into a Message type
func Unmarshal(s string) (Message, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, ErrMalformedMessage
	}

	// Remove any trailing newlines
	s = strings.Split(s, "\n")[0]
	parts := strings.Split(s, ",")

	if len(parts) == 0 {
		return nil, ErrMalformedMessage
	}

	switch parts[0] {
	case config.MsgDiscover:
		if len(parts) < 2 {
			return &Discover{List: []string{}}, nil
		}
		return &Discover{List: parts[1:]}, nil

	case config.MsgGet:
		if len(parts) < 2 {
			return nil, fmt.Errorf("%w: Get message requires file name", ErrMalformedMessage)
		}
		return &Get{Name: parts[1]}, nil

	case config.MsgFile:
		if len(parts) < 3 {
			return nil, fmt.Errorf("%w: File message requires method and port", ErrMalformedMessage)
		}

		method, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidMethod, err)
		}

		port, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidPort, err)
		}

		if method == config.TransferMethodTCP {
			return &File{Method: method, TCPPort: port}, nil
		}
		return &File{Method: method, UDPPort: port}, nil

	case config.MsgStopWait:
		return &StopWait{}, nil

	case config.MsgAsk:
		if len(parts) < 2 {
			return nil, fmt.Errorf("%w: Ask message requires file name", ErrMalformedMessage)
		}
		return &AskFile{Name: parts[1]}, nil

	case config.MsgBuffer:
		if len(parts) < 3 {
			return nil, fmt.Errorf("%w: Buffer message requires data", ErrMalformedMessage)
		}
		part, err := base64.StdEncoding.DecodeString(parts[2])
		if err != nil {
			return nil, fmt.Errorf("failed to decode buffer: %w", err)
		}
		return &Segment{Part: part}, nil

	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownMessage, parts[0])
	}
}

// ReliableUDPUnmarshal parses a reliable UDP protocol message
func ReliableUDPUnmarshal(s string) (Message, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, ErrMalformedMessage
	}

	s = strings.Split(s, "\n")[0]
	parts := strings.Split(s, ",")

	if len(parts) == 0 {
		return nil, ErrMalformedMessage
	}

	switch parts[0] {
	case message.Size:
		if len(parts) < 2 {
			return nil, fmt.Errorf("%w: Size message requires size value", ErrMalformedMessage)
		}
		size, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidSize, err)
		}
		return &Size{Size: size}, nil

	case message.FileName:
		if len(parts) < 2 {
			return nil, fmt.Errorf("%w: FileName message requires name", ErrMalformedMessage)
		}
		return &FileName{Name: parts[1]}, nil

	case message.Segment:
		if len(parts) < 2 {
			return nil, fmt.Errorf("%w: Segment message requires data", ErrMalformedMessage)
		}
		part, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return nil, fmt.Errorf("failed to decode segment: %w", err)
		}
		return &Segment{Part: part}, nil

	case config.MsgGet:
		if len(parts) < 2 {
			return nil, fmt.Errorf("%w: Get message requires file name", ErrMalformedMessage)
		}
		return &Get{Name: parts[1]}, nil

	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownMessage, parts[0])
	}
}
