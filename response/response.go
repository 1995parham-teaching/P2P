package response

import (
	"fmt"
	"strings"

	"github.com/elahe-dstn/p2p/message"
)

type Response interface {
	Marshal() string
	Unmarshal(string)
}

type File struct {
	Answer  bool
	TcpPort int
}

type Discover struct {
	List []string
}

func (f File) Marshal() string {
	ans := "n"

	if f.Answer {
		ans = "y"
	}

	return fmt.Sprintf("%s,%s,%d\n", message.FILE, ans, f.TcpPort)
}

func (d *Discover) Marshal() string {
	list := strings.Join(d.List, ",")

	return fmt.Sprintf("%s,%s\n", message.Discover, list)
}

func (d *Discover) Unmarshal(s string) {

}

func Unmarshal(s string) Response {
	s = strings.Split(s, "\n")[0]
	t := strings.Split(s, ",")

	switch t[0] {
	case message.Discover:
		return &Discover{List: t[1:]}
	}

	return nil
}
