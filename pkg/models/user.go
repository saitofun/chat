package models

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/saitofun/chat/pkg/depends/protoc"
	"github.com/saitofun/chat/pkg/errors"
	"github.com/saitofun/qlib/net/qsock"
)

type User struct {
	Name      string    `json:"name"`      // Name username global unique
	CreatedAt time.Time `json:"createdAt"` // CreatedAt when user created
	LastLogin time.Time `json:"lastLogin"` // LastLogin user last login at
	LogoffAt  time.Time `json:"logoffAt"`  // LogoffAt user logoff
}

func (u User) OnlineDuration() time.Duration {
	if u.LastLogin.Before(u.LogoffAt) {
		return u.LogoffAt.Sub(u.LastLogin)
	} else {
		return time.Since(u.LastLogin)
	}
}

type UserInfo struct {
	*User
	room   *Room
	node   *qsock.Node
	mtx    *sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
}

func NewUserInfo(user *User, node *qsock.Node) *UserInfo {
	user.LastLogin = time.Now()
	return &UserInfo{
		User: user,
		room: nil,
		node: node,
		mtx:  &sync.Mutex{},
	}
}

func (u *UserInfo) Pub(msg *protoc.Echo) error {
	u.mtx.Lock()
	defer u.mtx.Unlock()
	if u.room == nil {
		return errors.ErrNotEnterRoom
	}
	u.room.Pub(msg)
	return nil
}

func (u *UserInfo) EntryRoom(room *Room) {
	u.mtx.Lock()
	defer u.mtx.Unlock()

	if u.cancel != nil {
		u.cancel()
	}
	if u.room != nil {
		u.room.Leave(u.Name)
	}
	u.room = room
	u.ctx, u.cancel = context.WithCancel(context.Background())
	go u.consuming()
}

func (u *UserInfo) Leave() {
	u.mtx.Lock()
	defer u.mtx.Unlock()
	if u.room != nil {
		u.room.Leave(u.Name)
	}
}

func (u *UserInfo) Logoff() {
	u.mtx.Lock()
	defer u.mtx.Unlock()

	u.node.Stop()
	u.LogoffAt = time.Now()
	if u.cancel != nil {
		u.cancel()
	}
	if u.room != nil {
		u.room.Leave(u.Name)
	}
}

func (u *UserInfo) consuming() {
	cache, ch := u.room.Entry(u.Name)
	for _, msg := range cache {
		if msg.From == u.Name {
			continue
		}
		if err := u.node.WriteMessage(msg); err != nil {
			u.Logoff()
			return
		}
	}
	for {
		select {
		case <-u.ctx.Done():
			return
		case msg := <-ch:
			if msg.From == u.Name {
				continue
			}
			if err := u.node.WriteMessage(msg); err != nil {
				u.Logoff()
				return
			}
		}
	}
}

func (u *UserInfo) String() string {
	u.mtx.Lock()
	room := "未进入房间"
	if u.room != nil {
		room = strconv.Itoa(u.room.Id)
	}
	u.mtx.Unlock()

	du := u.OnlineDuration()

	return fmt.Sprintf("\n用户名: %s\n"+
		"登陆时间: %s\n"+
		"所在房间: %s\n"+
		"在线时长: %s\n", u.Name, u.LastLogin.Format("2006-01-02 15:04:05"),
		room, du.String())
}
