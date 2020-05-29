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
	TcpPort int
}

func (d *Discover) Marshal() string {
	list := strings.Join(d.List, ",")

	return fmt.Sprintf("%s,%s\n", Disco, list)
}

func (g *Get) Marshal() string {
	return fmt.Sprintf("%s,%s\n", G, g.Name)
}

func (f *File) Marshal() string {
	return fmt.Sprintf("%s,%d\n", F, f.TcpPort)
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
		return &File{TcpPort: port}
	}

	return nil
}
