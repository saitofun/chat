package protoc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"strconv"

	"github.com/saitofun/qlib/net/qbuf"
	"github.com/saitofun/qlib/net/qmsg"
)

// Type 消息类型
type Type uint32

const (
	CmdUnknown  Type = iota
	CmdEcho          // 回显消息
	CmdInstruct      // GM指令消息
)

func (t Type) String() string {
	switch t {
	case CmdEcho:
		return "ECHO"
	case CmdInstruct:
		return "INSTRUCT"
	default:
		return ""
	}
}

func (t Type) Uint32() uint32 {
	return uint32(t)
}

// Seq 消息序列号, 一个TCP连接内唯一
type Seq uint32

func (s Seq) String() string {
	return strconv.FormatUint(uint64(s), 10)
}

func (s Seq) Uint32() uint32 {
	return uint32(s)
}

// Header 消息头
type Header struct {
	Seq
	Type        // Cmd 消息类型
	Len  uint32 // Len 消息长度
}

func (h *Header) ID() qmsg.ID { return h.Seq }

func (h *Header) Bytes() []byte {
	ret := make([]byte, 0, 12)
	binary.BigEndian.PutUint32(ret[0:4], h.Seq.Uint32())
	binary.BigEndian.PutUint32(ret[8:12], h.Type.Uint32())
	binary.BigEndian.PutUint32(ret[4:8], h.Len)
	return ret
}

func (h Header) Marshal() ([]byte, error) { return h.Bytes(), nil }

func (h *Header) Unmarshal(d []byte) error {
	if len(d) < 12 {
		return fmt.Errorf("data lack")
	}
	h.Seq = Seq(binary.BigEndian.Uint32(d[0:4]))
	h.Type = Type(binary.BigEndian.Uint32(d[4:8]))
	h.Len = binary.BigEndian.Uint32(d[8:12])
	return nil
}

// Echo srv <-> cli
type Echo struct {
	Header
	Bodies []EchoMessage
}

var _ qmsg.Message = (*Echo)(nil)

func (m *Echo) Type() qmsg.Type { return CmdEcho }

func (m *Echo) Bytes() []byte {
	buf := bytes.NewBuffer(nil)

	buf.Write(m.Header.Bytes())
	for i := range m.Bodies {
		buf.Write(m.Bodies[i].Bytes())
	}

	return buf.Bytes()
}

func (m Echo) Marshal() ([]byte, error) { return m.Bytes(), nil }

func (m *Echo) Unmarshal(dat []byte) error {
	offset := 0
	if err := m.Header.Unmarshal(dat); err != nil {
		return err
	}
	offset += 12

	for cur := dat[offset:]; len(cur) > 0; {
		if len(cur) < 4 {
			return fmt.Errorf("data lack")
		}
		body := EchoMessage{
			Len: binary.BigEndian.Uint32(cur[0:4]),
		}
		offset += 4
		cur = cur[offset:]
		if uint32(len(cur)) < body.Len {
			return fmt.Errorf("data lack")
		}
		offset += int(body.Len)
		body.Body = string(cur[0:offset])
		m.Bodies = append(m.Bodies, body)
	}
	return nil
}

func NewEcho(seq Seq, bodies ...string) *Echo {
	ret := &Echo{
		Header: Header{Seq: seq, Type: CmdEcho},
		Bodies: make([]EchoMessage, len(bodies)),
	}
	for i := range bodies {
		if len(bodies[i]) == 0 {
			continue
		}
		ret.Bodies = append(ret.Bodies, EchoMessage{
			Len:  uint32(len(bodies[i])),
			Body: bodies[i],
		})
		ret.Header.Len += uint32(4 + len(bodies[i]))
	}
	return ret
}

type EchoMessage struct {
	Len  uint32
	Body string
}

func (m *EchoMessage) Bytes() []byte {
	buf := bytes.NewBuffer(nil)

	binary.Write(buf, binary.BigEndian, m.Len)
	buf.WriteString(m.Body)

	return buf.Bytes()
}

// Instruct cli -> srv
type Instruct struct {
	Header
	GmCmd
	ArgLen uint32 // ArgLen 参数长度
	Arg    string // Arg 参数内容
}

var _ qmsg.Message = (*Instruct)(nil)

func (m *Instruct) Type() qmsg.Type { return CmdInstruct }

func (m *Instruct) Bytes() []byte {
	buf := bytes.NewBuffer(nil)

	buf.Write(m.Header.Bytes())
	binary.Write(buf, binary.BigEndian, m.GmCmd)
	binary.Write(buf, binary.BigEndian, m.ArgLen)
	buf.WriteString(m.Arg)

	return buf.Bytes()
}

func (m Instruct) Marshal() ([]byte, error) { return m.Bytes(), nil }

func (m *Instruct) Unmarshal(dat []byte) error {
	offset := 0
	if err := m.Header.Unmarshal(dat); err != nil {
		return err
	}
	offset += 12
	if len(dat[offset:]) < 4 {
		return fmt.Errorf("data lack")
	}
	m.GmCmd = GmCmd(binary.BigEndian.Uint32(dat[offset : offset+4]))
	offset += 4
	if len(dat[offset:]) < 4 {
		return fmt.Errorf("data lack")
	}
	m.ArgLen = binary.BigEndian.Uint32(dat[offset : offset+4])
	offset += 4
	if uint32(len(dat[offset:])) < m.ArgLen {
		return fmt.Errorf("data lack")
	}
	if len(dat[offset:]) > 0 {
		m.Arg = string(dat[offset : offset+int(m.ArgLen)])
	}
	return nil
}

func NewInstruct(seq Seq, cmd GmCmd, args ...string) *Instruct {
	arglen, arg := uint32(0), ""
	if len(args) > 0 && args[0] != "" {
		arglen = uint32(len(args))
		arg = args[0]
	}
	return &Instruct{
		Header: Header{
			Seq:  seq,
			Type: CmdInstruct,
			Len:  uint32(8 + len(arg)),
		},
		GmCmd:  cmd,
		ArgLen: arglen,
		Arg:    arg,
	}
}

type GmCmd uint32

const (
	GmCmdUnknown GmCmd = iota
	GmCreateUser
	GmLogin
	GmRoomList
	GmEnterRoom
	GmStats
	GmPopular
)

type parser struct{}

var Parser qmsg.Parser = &parser{}

func (p parser) Marshal(buf qbuf.Buffer, msg qmsg.Message) error {
	var err error

	buf.Reset()

	switch msg.(type) {
	case *Echo:
		_, err = buf.Write(buf.Bytes())
	case *Instruct:
		_, err = buf.Write(buf.Bytes())
	default:
		err = fmt.Errorf("unknown message: %s", reflect.TypeOf(msg))
	}
	return err
}

func (p parser) Unmarshal(buf qbuf.Buffer) (qmsg.Message, error) {
	tmp, err := buf.Probe(12)
	if err != nil {
		return nil, err
	}

	header := &Header{}
	if err = header.Unmarshal(tmp); err != nil {
		return nil, err
	}

	if _, err = buf.Probe(int(header.Len)); err != nil {
		return nil, err
	}

	dat := make([]byte, int(header.Len)+12)
	_, _ = buf.Read(dat)

	var msg qmsg.Message

	switch header.Type {
	case CmdEcho:
		echo := &Echo{}
		if err = echo.Unmarshal(dat); err != nil {
			msg = echo
		}
	case CmdInstruct:
		instruct := &Instruct{}
		if err = instruct.Unmarshal(dat); err != nil {
			msg = instruct
		}
	default:
		err = fmt.Errorf("unknown message: %d", header.Type)
	}
	if err != nil {
		return nil, err
	}
	return msg, nil
}
