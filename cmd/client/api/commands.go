package api

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/saitofun/chat/cmd/config"
	"github.com/saitofun/chat/pkg/depends/protoc"
	"github.com/saitofun/chat/pkg/models"
	"github.com/saitofun/qlib/encoding/qjson"
	"github.com/saitofun/qlib/net/qsock"
)

var (
	client      *qsock.Client
	User        *models.User
	LoginAt     *time.Time
	mtx         = &sync.Mutex{}
	ctx, cancel = context.WithCancel(context.Background())
)

func init() {
	var err error
	client, err = qsock.NewClient(
		qsock.ClientOptionParser(protoc.Parser),
		qsock.ClientOptionRemote(config.ClientAddr),
		qsock.ClientOptionProtocol(qsock.ProtocolTCP),
		qsock.ClientOptionNodeID(" client"),
		// qsock.ClientOptionRoute(protoc.CmdEcho, func(ev *qsock.Event) {
		// 	Output(ev.Payload())
		// }),
	)
	if err != nil {
		panic(err)
	}
	go receiving()
	go handling()
}

func Logoff(reason ...interface{}) {
	if client != nil && !client.IsClosed() {
		client.Close(reason...)
	}
}

func Shutdown() {
	cancel()
	os.Exit(-1)
}

func receiving() {
	defer Shutdown()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		msg, err := client.RecvMessage()
		if err != nil {
			if err == qsock.ENodeTimeout {
				continue
			}
			Output("[ERROR]: " + err.Error())
			Output(err.Error())
			return
		}
		if msg != nil {
			Output(msg)
		}
	}
}

// HandleInput 用户输入处理
func HandleInput() interface{} {
	reader := bufio.NewReader(os.Stdin)
	bytes, _, err := reader.ReadLine()
	if err != nil {
		return err
	}
	if len(bytes) == 0 {
		return ""
	}
	line := string(bytes)
	if line[0] != '/' {
		from := ""
		if User != nil {
			from = User.Name
		}
		return handlePub(from, line)
	}
	words := strings.Split(line[1:], " ")
	cmd, arg := "", ""
	if len(words) > 0 {
		cmd = strings.TrimSpace(words[0])
	}
	if len(words) > 1 {
		arg = strings.TrimSpace(words[1])
	}
	return HandleCommand(cmd, arg)
}

func handlePub(from, line string) interface{} {
	err := client.SendMessage(protoc.NewEcho(
		protoc.Seq(uuid.New().ID()),
		from,
		line,
	))
	if err != nil {
		return err
	}
	return "> "
}

// HandleCommand 命令请求
func HandleCommand(cmd string, arg ...string) interface{} {
	switch cmd {
	case "reg":
		if len(arg) == 0 || arg[0] == "" {
			return "请输入用户名"
		}
		return handleRegister(arg[0])
	case "login":
		if len(arg) == 0 || arg[0] == "" {
			return "请输入用户名"
		}
		return handleLogin(arg[0])
	case "rooms":
		return handleRoomList()
	case "room":
		if len(arg) == 0 || arg[0] == "" {
			return "请输入房间号"
		}
		if id, err := strconv.Atoi(arg[0]); err != nil {
			return "房间号非法"
		} else {
			return handleEnterRoom(id)
		}
	case "stats":
		if len(arg) == 0 || arg[0] == "" {
			return "请输入用户名"
		}
		return handleUserInfo(arg[0])
	case "popular":
		if len(arg) == 0 || arg[0] == "" {
			return "请输入房间号"
		}
		if id, err := strconv.Atoi(arg[0]); err != nil {
			return "房间号非法"
		} else {
			return handlePopularWords(id)
		}
	default:
		return "无效命令"
	}
}

func handleRegister(username string) interface{} {
	rsp, err := client.Request(protoc.NewInstruct(
		protoc.Seq(uuid.New().ID()),
		protoc.GmCreateUser,
		username,
	))
	if err != nil {
		return err
	}
	return rsp
}

func handleLogin(username string) interface{} {
	rsp, err := client.Request(protoc.NewInstruct(
		protoc.Seq(uuid.New().ID()),
		protoc.GmLogin,
		username,
	))
	if err != nil {
		return err
	}
	return rsp
}

func handleRoomList() interface{} {
	rsp, err := client.Request(protoc.NewInstruct(
		protoc.Seq(uuid.New().ID()),
		protoc.GmRoomList,
	))
	if err != nil {
		return err
	}
	return rsp
}

func handleEnterRoom(id int) interface{} {
	rsp, err := client.Request(protoc.NewInstruct(
		protoc.Seq(uuid.New().ID()),
		protoc.GmEnterRoom,
		strconv.Itoa(id),
	))
	if err != nil {
		return err
	}
	return rsp
}

func handleUserInfo(args ...string) interface{} {
	rsp, err := client.Request(protoc.NewInstruct(
		protoc.Seq(uuid.New().ID()),
		protoc.GmStats,
		args...,
	))
	if err != nil {
		return err
	}
	return rsp
}

func handlePopularWords(id int) interface{} {
	rsp, err := client.Request(protoc.NewInstruct(
		protoc.Seq(uuid.New().ID()),
		protoc.GmPopular,
		strconv.Itoa(id),
	))
	if err != nil {
		return err
	}
	return rsp
}

// handling handle user input
func handling() {
	fmt.Print("> ")
	for {
		Output(HandleInput())
	}
}

func Output(msg interface{}) {
	mtx.Lock()
	defer mtx.Unlock()
	if msg == nil {
		return
	}
	body := ""
	if err, ok := msg.(error); ok && err != nil {
		body = "[系统错误] " + err.Error()
	} else if s, ok := msg.(interface{ String() string }); ok {
		body = s.String()
	} else if s, ok := msg.(string); ok {
		body = s
	} else {
		body = qjson.UnsafeMarshalString(msg)
	}
	fmt.Printf(body)
	fmt.Print("\n> ")
}
