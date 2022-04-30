package api

import (
	"strconv"

	"github.com/saitofun/chat/pkg/depends/protoc"
	"github.com/saitofun/chat/pkg/errors"
	"github.com/saitofun/chat/pkg/modules/rooms"
	"github.com/saitofun/chat/pkg/modules/users"
	"github.com/saitofun/qlib/encoding/qjson"
	"github.com/saitofun/qlib/net/qsock"
)

func OnEcho(msg *protoc.Echo, c *qsock.Node) error {
	ctrl := users.Controller()
	user := ctrl.GetByClientID(c.ID())
	if user == nil {
		return errors.ErrUserNotLogin
	}
	if err := user.Pub(msg); err != nil {
		return err
	}
	return nil
}

func OnGmCmd(msg *protoc.Instruct, c *qsock.Node) error {
	var (
		err      error
		ctrlUser = users.Controller()
		ctrlRoom = rooms.Controller()
		user     = ctrlUser.GetByClientID(c.ID())
		rsp      interface{}
	)

	if user == nil && msg.GmCmd != protoc.GmCreateUser && msg.GmCmd != protoc.GmLogin {
		return errors.ErrUserNotLogin
	}
	switch msg.GmCmd {
	case protoc.GmCreateUser:
		rsp, err = ctrlUser.CreateUser(msg.Arg, c)
	case protoc.GmLogin:
		rsp, err = ctrlUser.UserLogin(msg.Arg, c)
	case protoc.GmRoomList:
		rsp = ctrlRoom.RoomList()
	case protoc.GmEnterRoom:
		var roomID int
		roomID, err = strconv.Atoi(msg.Arg)
		if err != nil {
			err = errors.ErrInvalidRoomID
		} else {
			room := ctrlRoom.GetByID(roomID)
			if room == nil {
				ctrlRoom.CreateRoom()
			}
			user.EntryRoom(room)
			rsp = room
		}
	case protoc.GmStats:
		rsp = user
	case protoc.GmPopular:
		var roomID int
		roomID, err = strconv.Atoi(msg.Arg)
		if err != nil {
			err = errors.ErrInvalidRoomID
		} else {
			room := ctrlRoom.GetByID(roomID)
			if room == nil {
				return errors.ErrRoomIDNotExists
			}
			rsp = room.PopularWords()
		}
	default:
		err = errors.ErrUnknownGmCmd
	}

	if err != nil {
		return err
	}
	body := ""
	if s, ok := rsp.(interface{ String() string }); ok {
		body = s.String()
	} else {
		body = qjson.UnsafeMarshalString(rsp)
	}
	return c.SendMessage(protoc.NewEcho(msg.Seq, body))
}
