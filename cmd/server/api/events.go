package api

import (
	"fmt"
	"strconv"

	"github.com/saitofun/chat/pkg/depends/protoc"
	"github.com/saitofun/chat/pkg/errors"
	"github.com/saitofun/chat/pkg/models"
	"github.com/saitofun/chat/pkg/modules/rooms"
	"github.com/saitofun/chat/pkg/modules/users"
	"github.com/saitofun/qlib/encoding/qjson"
	"github.com/saitofun/qlib/net/qsock"
)

func OnEcho(ev *qsock.Event) {
	msg := ev.Payload().(*protoc.Echo)
	c := ev.Node()
	ctrl := users.Controller()
	user := ctrl.GetByClientID(c.ID())
	if user == nil {
		Response(msg.Seq, errors.ErrUserNotLogin, c)
		return
	}
	msg.SetFrom(user.Name)
	if err := user.Pub(msg); err != nil {
		Response(msg.Seq, err, c)
	}
}

func OnGmCmd(ev *qsock.Event) {
	var (
		msg, c   = ev.Payload().(*protoc.Instruct), ev.Node()
		ctrlUser = users.Controller()
		ctrlRoom = rooms.Controller()
		user     = ctrlUser.GetByClientID(c.ID())
		seq      = msg.Seq
	)

	if user == nil && msg.GmCmd != protoc.GmCreateUser && msg.GmCmd != protoc.GmLogin {
		Response(seq, errors.ErrUserNotLogin, c)
		return
	}
	switch msg.GmCmd {
	case protoc.GmCreateUser:
		info, err := ctrlUser.CreateUser(msg.Arg, c)
		if err != nil {
			Response(seq, err, c)
			return
		}
		Response(seq, info, c)
		return
	case protoc.GmLogin:
		u, err := ctrlUser.UserLogin(msg.Arg, c)
		if err != nil {
			Response(seq, err, c)
			return
		}
		Response(seq, u, c)
		return
	case protoc.GmRoomList:
		Response(seq, ctrlRoom.RoomList(), c)
		return
	case protoc.GmEnterRoom:
		roomID, err := strconv.Atoi(msg.Arg)
		if err != nil {
			Response(seq, errors.ErrInvalidRoomID, c)
			return
		}
		room := ctrlRoom.GetByID(roomID)
		if room == nil {
			if room, err = ctrlRoom.CreateRoom(roomID); err != nil {
				Response(seq, errors.ErrInvalidRoomID, c)
				return
			}
		}
		user.EntryRoom(room)
		Response(seq, room, c)
		return
	case protoc.GmStats:
		u := ctrlUser.GetByName(msg.Arg)
		if u == nil {
			Response(seq, errors.ErrUserNotExisted, c)
			return
		}
		i := ctrlUser.GetUserInfoByName(msg.Arg)
		if i == nil {
			i = &models.UserInfo{User: u}
			return
		}
		Response(seq, i, c)
		return
	case protoc.GmPopular:
		roomID, err := strconv.Atoi(msg.Arg)
		if err != nil {
			Response(seq, errors.ErrInvalidRoomID, c)
			return
		}
		room := ctrlRoom.GetByID(roomID)
		if room == nil {
			Response(seq, errors.ErrRoomIDNotExists, c)
			return
		}
		Response(seq, room.PopularWords(), c)
		return
	default:
		Response(seq, errors.ErrUnknownGmCmd, c)
		return
	}
}

func Response(seq protoc.Seq, msg interface{}, c *qsock.Node) {
	body := ""
	if err, ok := msg.(error); ok {
		body = "[SERVER] " + err.Error()
	} else if s, ok := msg.(interface{ String() string }); ok {
		body = s.String()
	} else {
		body = qjson.UnsafeMarshalString(msg)
	}
	if err := c.SendMessage(protoc.NewEcho(seq, "SYSTEM", body)); err != nil {
		fmt.Println(err)
	}
}
