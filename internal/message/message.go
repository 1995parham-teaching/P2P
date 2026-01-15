package message

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/1995parham-teaching/P2P/internal/config"
)

var (
	ErrMalformedMessage = errors.New("malformed message")
	ErrUnknownMessage   = errors.New("unknown message type")
	ErrInvalidPort      = errors.New("invalid port number")
	ErrInvalidMethod    = errors.New("invalid transfer method")
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
}

func (d *Discover) Marshal() string {
	list := strings.Join(d.List, ",")
	return fmt.Sprintf("%s,%s\n", config.MsgDiscover, list)
}

func (g *Get) Marshal() string {
	return fmt.Sprintf("%s,%s\n", config.MsgGet, g.Name)
}

func (f *File) Marshal() string {
	return fmt.Sprintf("%s,%d,%d\n", config.MsgFile, f.Method, f.TCPPort)
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

		return &File{Method: method, TCPPort: port}, nil

	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownMessage, parts[0])
	}
}
