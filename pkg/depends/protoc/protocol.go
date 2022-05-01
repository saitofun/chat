package protoc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"

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

func (t Type) Uint32() uint32 { return uint32(t) }

// Seq 消息序列号, 一个TCP连接内唯一
type Seq uint32

func (s Seq) String() string { return strconv.FormatUint(uint64(s), 10) }

func (s Seq) Uint32() uint32 { return uint32(s) }

// Header 消息头
type Header struct {
	Seq
	Type
	Len uint32 // Len 消息长度
}

func (h *Header) ID() qmsg.ID { return h.Seq }

func (h *Header) Bytes() []byte {
	ret := make([]byte, 12)
	order.PutUint32(ret[0:4], h.Seq.Uint32())
	order.PutUint32(ret[4:8], h.Type.Uint32())
	order.PutUint32(ret[8:12], h.Len)
	return ret
}

func (h Header) Marshal() ([]byte, error) { return h.Bytes(), nil }

func (h *Header) Unmarshal(d []byte) error {
	if len(d) < 12 {
		return fmt.Errorf("data lack")
	}
	h.Seq = Seq(order.Uint32(d[0:4]))
	h.Type = Type(order.Uint32(d[4:8]))
	h.Len = order.Uint32(d[8:12])
	return nil
}

// Echo srv <-> cli
type Echo struct {
	Header
	From string
	Body string
}

var _ qmsg.Message = (*Echo)(nil)

func (m *Echo) Type() qmsg.Type { return CmdEcho }

func (m *Echo) Bytes() []byte {
	buf := bytes.NewBuffer(nil)

	buf.Write(m.Header.Bytes())
	buf.Write(BinaryText(m.From))
	buf.Write(BinaryText(m.Body))

	return buf.Bytes()
}

func (m Echo) Marshal() ([]byte, error) { return m.Bytes(), nil }

func (m *Echo) Unmarshal(dat []byte) error {
	offset := uint32(0)
	if err := m.Header.Unmarshal(dat); err != nil {
		return err
	}
	offset += 12

	str, delta, err := ParseString(dat[offset:])
	if err != nil {
		return err
	}
	offset += delta
	m.From = str

	str, delta, err = ParseString(dat[offset:])
	if err != nil {
		return err
	}
	offset += delta
	m.Body = str

	return nil
}

func (m *Echo) SetFrom(from string) {
	m.From = from
	m.Len = uint32(len(m.From) + len(m.Body) + 8)
}

func (m *Echo) SetBody(body string) {
	m.From = body
	m.Len = uint32(len(m.From) + len(m.Body) + 8)
}

func (m *Echo) String() string { return fmt.Sprintf("[%s]: %s", m.From, m.Body) }

func NewEcho(seq Seq, from, body string) *Echo {
	return &Echo{
		Header: Header{
			Seq:  seq,
			Type: CmdEcho,
			Len:  uint32(8 + len(from) + len(body)),
		},
		From: from,
		Body: body,
	}
}

// Instruct cli -> srv
type Instruct struct {
	Header
	GmCmd
	Arg string // Arg 参数内容
}

var _ qmsg.Message = (*Instruct)(nil)

func (m *Instruct) Type() qmsg.Type { return CmdInstruct }

func (m *Instruct) Bytes() []byte {
	buf := bytes.NewBuffer(nil)

	buf.Write(m.Header.Bytes())
	binary.Write(buf, binary.BigEndian, m.GmCmd)
	buf.Write(BinaryText(m.Arg))

	return buf.Bytes()
}

func (m Instruct) Marshal() ([]byte, error) { return m.Bytes(), nil }

func (m *Instruct) Unmarshal(dat []byte) error {
	offset := uint32(0)
	if err := m.Header.Unmarshal(dat); err != nil {
		return err
	}
	offset += 12
	if uint32(len(dat[offset:])) != m.Len {
		return errUnexpectedPayloadLength
	}

	if len(dat[offset:]) < 4 {
		return errDataLack
	}
	m.GmCmd = GmCmd(binary.BigEndian.Uint32(dat[offset : offset+4]))
	offset += 4
	str, delta, err := ParseString(dat[offset:])
	if err != nil {
		return err
	}
	offset += delta
	m.Arg = str
	return nil
}

func (m *Instruct) String() string { return strings.TrimSpace(m.GmCmd.String() + " " + m.Arg) }

func NewInstruct(seq Seq, cmd GmCmd, args ...string) *Instruct {
	arg := ""
	if len(args) > 0 && args[0] != "" {
		arg = args[0]
	}
	return &Instruct{
		Header: Header{
			Seq:  seq,
			Type: CmdInstruct,
			Len:  uint32(8 + len(arg)),
		},
		GmCmd: cmd,
		Arg:   arg,
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

func (gm GmCmd) String() string {
	switch gm {
	case GmCreateUser:
		return "/reg"
	case GmLogin:
		return "/login"
	case GmRoomList:
		return "/rooms"
	case GmEnterRoom:
		return "/room"
	case GmStats:
		return "/stats"
	case GmPopular:
		return "/popular"
	default:
		return ""
	}
}

type parser struct{}

var Parser qmsg.Parser = &parser{}

func (p parser) Marshal(buf qbuf.Buffer, msg qmsg.Message) error {
	var err error

	buf.Reset()

	switch _msg := msg.(type) {
	case *Echo:
		_, err = buf.Write(_msg.Bytes())
	case *Instruct:
		_, err = buf.Write(_msg.Bytes())
	default:
		err = errUnknownMessage
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
			return nil, err
		}
		msg = echo
	case CmdInstruct:
		instruct := &Instruct{}
		if err = instruct.Unmarshal(dat); err != nil {
			return nil, err
		}
		msg = instruct
	default:
		return nil, errUnknownMessage
	}
	return msg, nil
}

func BinaryText(v string) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, len(v)+4))

	binary.Write(buf, order, uint32(len(v)))
	buf.WriteString(v)
	return buf.Bytes()
}

func ParseString(v []byte) (string, uint32, error) {
	if len(v) < 4 {
		return "", 0, errDataLack
	}
	length := order.Uint32(v[0:4])
	if uint32(len(v[4:])) < length {
		return "", 0, errDataLack
	}
	str := string(v[4 : length+4])
	return str, length + 4, nil
}

var (
	order = binary.BigEndian

	errDataLack                = errors.New("CHAT:data lack")
	errUnexpectedPayloadLength = errors.New("CHAT:unexpected payload length")
	errUnknownMessage          = errors.New("CHAT:unknown message")
)
