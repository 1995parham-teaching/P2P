package response

import (
	"fmt"

	"github.com/elahe-dstn/p2p/message"
)

type File struct {
	Answer bool
	TcpPort int
}

func (f File) Marshal() string {
	ans := "n"

	if f.Answer {
		ans = "y"
	}

	return fmt.Sprintf("%s,%s,%d\n", message.FILE, ans, f.TcpPort)
}
