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
	TCPPort int
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
	return fmt.Sprintf("%s,%d\n", F, f.TCPPort)
}

func (s *StopWait) Marshal() string {
	return fmt.Sprintf("%s\n", SW)
}

func (a *AskFile) Marshal() string {
	return fmt.Sprintf("%s,%s\n", Ask, a.Name)
}

func (s *Size) Marshal() string {
	fileSize := FileSize + ","
	fileSize += strconv.Itoa(s.Seq)
	fileSize += ","
	fileSize += strconv.FormatInt(s.Size, 10)
	fileSize += "\n"
	//fileSize = fillString(fileSize, 10)

	return fileSize
}

func (n *FileName) Marshal() string {
	fileName := Name + ","
	fileName += strconv.Itoa(n.Seq)
	fileName += ","
	fileName += n.Name
	fileName += "\n"
	//fileName = fillString(fileName, 64)

	return fileName
}

func (s *Segment) Marshal() string {
	return fmt.Sprintf("%s,%d,%s\n", Buffer, s.Seq, s.Part)
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
		port, _ := strconv.Atoi(t[1])
		return &File{TCPPort: port}
	case SW:
		return &StopWait{}
	case Ask:
		name := t[1]
		return &AskFile{Name:name}
	case FileSize:
		seq,_ := strconv.Atoi(t[1])
		size,_ := strconv.Atoi(t[2])
		size64 := int64(size)

		return &Size{
			Size: size64,
			Seq:  seq,
		}

	case Name:
		seq,_ := strconv.Atoi(t[1])
		name := t[2]

		return &FileName{
			Name: name,
			Seq:  seq,
		}
	case Buffer:
		seq,_ := strconv.Atoi(t[1])

		return &Segment{
			Part: []byte(t[2]),
			Seq:  seq,
		}
	case Ack:
		seq,_ := strconv.Atoi(t[1])

		return &Acknowledgment{Seq:seq}
	}

	return nil
}

func fillString(retunString string, toLength int) string {
	for {
		lengtString := len(retunString)
		if lengtString < toLength {
			retunString += ":"
			continue
		}

		break
	}

	return retunString
}
