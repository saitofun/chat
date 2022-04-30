package api

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/saitofun/chat/cmd/config"
	"github.com/saitofun/chat/pkg/depends/protoc"
	"github.com/saitofun/chat/pkg/models"
	"github.com/saitofun/qlib/net/qmsg"
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

func Pub(msg ...string) error {
	return client.SendMessage(protoc.NewEcho(
		protoc.Seq(uuid.New().ID()),
		msg...,
	))

}

func NewUser(username string) (qmsg.Message, error) {
	ret, err := client.Request(protoc.NewInstruct(
		protoc.Seq(uuid.New().ID()),
		protoc.GmCreateUser,
		username,
	))
	if err != nil {
		return nil, err
	}
	// @todo parse ret and update user info and login time
	return ret, nil
}

func Login(username string) (qmsg.Message, error) {
	ret, err := client.Request(protoc.NewInstruct(
		protoc.Seq(uuid.New().ID()),
		protoc.GmLogin,
		username,
	))
	if err != nil {
		return nil, err
	}
	// @todo parse ret and update user info and login time
	return ret, nil
}

func EnterRoom(id int) (qmsg.Message, error) {
	return client.Request(protoc.NewInstruct(
		protoc.Seq(uuid.New().ID()),
		protoc.GmEnterRoom,
		strconv.Itoa(id),
	))
}

func GetRoomList() (qmsg.Message, error) {
	return client.Request(protoc.NewInstruct(
		protoc.Seq(uuid.New().ID()),
		protoc.GmRoomList,
	))
}

func UserInfo(args ...string) (qmsg.Message, error) {
	return client.Request(protoc.NewInstruct(
		protoc.Seq(uuid.New().ID()),
		protoc.GmStats,
		args...,
	))
}

func Popular(id int) (qmsg.Message, error) {
	return client.Request(protoc.NewInstruct(
		protoc.Seq(uuid.New().ID()),
		protoc.GmPopular,
		strconv.Itoa(id),
	))
}

func Shutdown() { cancel() }

func receiving() {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		msg, err := client.RecvMessage()
		if err != nil {
			return
		}
		_ = msg
		// @todo echo msg and
	}
}

// handling handle user input
func handling() {
	var msg string
	for {
		fmt.Printf("> ")
		fmt.Scanln(&msg)

		if msg[0] == '/' {
			msg = msg[1:]
			args := strings.Split(msg, " ")
			cmd := ""
			arg := ""
			if len(args) == 1 {
				cmd = strings.TrimSpace(args[0])
			} else if len(args) > 1 {
				cmd = strings.TrimSpace(args[0])
				arg = strings.TrimSpace(args[1])
			}
			switch cmd {
			case "register":
				NewUser(arg)
			case "login":
				Login(arg)
			case "rooms":
				GetRoomList()
			case "room":
				id, err := strconv.Atoi(arg)
				if err != nil {
					fmt.Println("房间号非法")
					continue
				}
				EnterRoom(id)
			case "stats":
				if arg == "" && User != nil && User.Name != "" {
					Print("用户名非法")
					continue
				}
				UserInfo(arg)
			case "popular":
				id, err := strconv.Atoi(arg)
				if err != nil {
					Print("房间号非法")
					continue
				}
				Popular(id)
			default:
				Print("非法指令")
			}
		} else {
			Pub(msg)
		}
	}
}

func Print(s string) {
	mtx.Lock()
	defer mtx.Unlock()
	fmt.Println(s)
}
