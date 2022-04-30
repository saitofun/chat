package models

import (
	"sync"

	"github.com/saitofun/chat/cmd/config"
	"github.com/saitofun/qlib/container/qtype"
)

type Room struct {
	Id       int
	mq       chan *UserMessage
	populars interface{}
	mtx      *sync.Mutex
	users    *qtype.Int
}

func NewRoom(id int) *Room {
	return &Room{
		Id:    id,
		mq:    make(chan *UserMessage, config.MaxRoomCache),
		users: qtype.NewInt(),
		mtx:   &sync.Mutex{},
	}
}

// Pub 用户发布消息
func (r *Room) Pub(msg *UserMessage) {
	r.mq <- msg
	// @todo update populars
}

// Entry 用户进入房间
func (r *Room) Entry() <-chan *UserMessage {
	r.users.Add(1)
	return r.mq
}

// Leave 用户离开
func (r *Room) Leave() { r.users.Add(-1) }

// UserCount 房间用户数
func (r *Room) UserCount() int { return r.users.Val() }

// PopularWords 房间频率最高的词
func (r *Room) PopularWords() []string { return nil }
