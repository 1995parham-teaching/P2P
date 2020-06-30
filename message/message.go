package message

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/elahe-dastan/reliable_UDP/message"
)

const (
	Disco = "DISCOVER"
	G     = "Get"
	F     = "File"
	SW	  = "SW"
	Ask   = "Ask"
	FileSize = "FSize"
	Name     = "Name"
	Buffer   = "buffer"
	BUFFERSIZE = 1024
	Ack        = "Ack"
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
	Method int
	TCPPort int
	UDPPort int
}

type StopWait struct {

}

type AskFile struct {
	Name  string
}

type Size struct {
	Size int64
}

type FileName struct{
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

	return fmt.Sprintf("%s,%s\n", Disco, list)
}

func (g *Get) Marshal() string {
	return fmt.Sprintf("%s,%s\n", G, g.Name)
}

func (f *File) Marshal() string {
	// method 1 means TCP
	if f.Method == 1 {
		return fmt.Sprintf("%s,%d,%d\n", F, f.Method, f.TCPPort)
	}else {
		// reliable tcp
		return fmt.Sprintf("%s,%d,%d\n", F, f.Method, f.UDPPort)
	}
}

func (s *StopWait) Marshal() string {
	return fmt.Sprintf("%s\n", SW)
}

func (a *AskFile) Marshal() string {
	return fmt.Sprintf("%s,%s\n", Ask, a.Name)
}

func (a *Acknowledgment) Marshal() string {
	return fmt.Sprintf("%s,%d\n", Ack, a.Seq)
}

func (s *Size) Marshal() string {
	fileSize := message.Size + "," +
		strconv.FormatInt(s.Size, 10) + "\n"

	return fileSize
}

func (n *FileName) Marshal() string {
	fileName := message.FileName + "," + n.Name + "\n"

	return fileName
}

func (s *Segment) Marshal() string {
	return fmt.Sprintf("%s,%s\n", message.Segment, base64.StdEncoding.EncodeToString(s.Part))
}

func Unmarshal(s string) Message {
	s = strings.Split(s, "\n")[0]
	t := strings.Split(s, ",")

	switch t[0] {
	case Disco:
		return &Discover{List: t[1:]}
	case G:
		return &Get{Name: t[1]}
	case F:
		method, _ := strconv.Atoi(t[1])
		port, _ := strconv.Atoi(t[2])

		if method == 1 {
			return &File{TCPPort: port}
		}else {
			return &File{UDPPort:port}
		}

	case SW:
		return &StopWait{}
	case Ask:
		name := t[1]
		return &AskFile{Name:name}

	case Buffer:
		part, _ := base64.StdEncoding.DecodeString(t[2])

		return &Segment{
			Part: part,
		}
	}

	return nil
}


func ReliableUDPUnmarshal(s string) Message {
	s = strings.Split(s, "\n")[0]
	t := strings.Split(s, ",")

	switch t[0] {
	case message.Size:
		size, _ := strconv.Atoi(t[1])
		size64 := int64(size)

		return &Size{
			Size: size64,
		}
	case message.FileName:
		name := t[1]

		return &FileName{
			Name: name,
		}
	case message.Segment:
		part, _ := base64.StdEncoding.DecodeString(t[1])

		return &Segment{
			Part: part,
		}
	case G:
		return &Get{Name: t[1]}
	}

	return nil
}