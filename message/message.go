package message

import (
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
	TCPPort int
}

type StopWait struct {

}

type AskFile struct {
	Name  string
}

func (d *Discover) Marshal() string {
	list := strings.Join(d.List, ",")

	return fmt.Sprintf("%s,%s\n", Disco, list)
}

func (g *Get) Marshal() string {
	return fmt.Sprintf("%s,%s\n", G, g.Name)
}

func (f *File) Marshal() string {
	return fmt.Sprintf("%s,%d\n", F, f.TCPPort)
}

func (s *StopWait) Marshal() string {
	return fmt.Sprintf("%s\n", SW)
}

func (a *AskFile) Marshal() string {
	return fmt.Sprintf("%s,%s\n", Ask, a.Name)
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
		port, _ := strconv.Atoi(t[1])
		return &File{TCPPort: port}
	case SW:
		return &StopWait{}
	}

	return nil
}
