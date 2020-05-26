package request

import (
	"fmt"
	"strings"

	"github.com/elahe-dstn/p2p/message"
)

type Request interface {
	Marshal() string
}

type Discover struct {

}

type File struct {
	Name string
}

func (d Discover) Marshal() string {
	return fmt.Sprintf("%s\n", message.Discover)
}

func (f File) Marshal() string {
	return fmt.Sprintf("%s,%s\n", message.FILE, f.Name)
}

func Unmarshal(req string) Request {
	arr := strings.Split(req, ",")
	t := strings.TrimSpace(arr[0])

	switch t {
	case message.Discover:
		return Discover{}
	case message.FILE:
		return File{}
	}
}
