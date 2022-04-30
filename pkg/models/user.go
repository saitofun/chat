package models

import (
	"context"
	"sync"
	"time"

	"github.com/saitofun/chat/pkg/depends/protoc"
	"github.com/saitofun/chat/pkg/errors"
	"github.com/saitofun/qlib/container/qtype"
	"github.com/saitofun/qlib/net/qmsg"
	"github.com/saitofun/qlib/net/qsock"
)

type UserMessage struct {
	User string       // User whom send this
	Body qmsg.Message // Body message body
}

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
	sub    *qtype.Int
	node   *qsock.Node
	mtx    *sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
}

func NewUserInfo(user *User, node *qsock.Node) *UserInfo {
	user.LastLogin = time.Now()
	return &UserInfo{
		User:   user,
		room:   nil,
		sub:    qtype.NewInt(),
		node:   node,
		mtx:    &sync.Mutex{},
		cancel: nil,
	}
}

func (u *UserInfo) Pub(msg *protoc.Echo) error {
	u.mtx.Lock()
	defer u.mtx.Unlock()
	if u.room == nil {
		return errors.ErrNotEnterRoom
	}
	// @todo filter dirty word
	u.room.Pub(&UserMessage{
		User: u.Name,
		Body: msg,
	})
	return nil
}

func (u *UserInfo) EntryRoom(room *Room) {
	u.mtx.Lock()
	defer u.mtx.Unlock()

	if u.sub.Val() != 0 {
		u.cancel()
		u.sub.Set(0)
	}
	u.room = room
	u.ctx, u.cancel = context.WithCancel(context.Background())
	u.room.Entry()
	go u.consuming()
}

func (u *UserInfo) Logoff() {
	u.mtx.Lock()
	defer u.mtx.Unlock()

	u.node.Stop()
	u.LogoffAt = time.Now()
	if u.room != nil {
		u.room.Leave()
	}
}

func (u *UserInfo) consuming() {
	defer u.Logoff()
	ch := u.room.Entry()
	for {
		select {
		case <-u.ctx.Done():
			return
		case msg := <-ch:
			if msg.User == u.Name {
				continue
			}
			if err := u.node.WriteMessage(msg.Body); err != nil {
				return
			}
		}
	}
}
