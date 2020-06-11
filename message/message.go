package message

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
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
	Seq  int
}

type FileName struct{
	Name string
	Seq int
}

type Segment struct {
	Part []byte
	Seq int
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

func (s *Segment) Marshal() string {
	return fmt.Sprintf("%s,%d,%s\n", Buffer, s.Seq, base64.StdEncoding.EncodeToString(s.Part))
}

func (a *Acknowledgment) Marshal() string {
	return fmt.Sprintf("%s,%d\n", Ack, a.Seq)
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
		seq,_ := strconv.Atoi(t[1])
		part, _ := base64.StdEncoding.DecodeString(t[2])

		return &Segment{
			Part: part,
			Seq:  seq,
		}
	}

	return nil
}
