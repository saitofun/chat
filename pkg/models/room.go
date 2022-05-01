package models

import (
	"fmt"
	"log"
	"sync"

	"github.com/saitofun/chat/cmd/config"
	"github.com/saitofun/chat/pkg/depends/protoc"
	"github.com/saitofun/qlib/container/qlist"
)

type Rooms []Room

func (rs Rooms) String() string {
	ret := "\n"
	for i := range rs {
		ret += fmt.Sprintf("房间号: %-3d 当前在线用户: %d\n", rs[i].Id, rs[i].UserCount())
	}
	return ret
}

type Room struct {
	Id       int
	mq       chan *protoc.Echo
	populars interface{}
	mtx      *sync.Mutex
	users    map[string]chan *protoc.Echo
	cache    *qlist.List
}

func NewRoom(id int) *Room {
	return &Room{
		Id:    id,
		mq:    make(chan *protoc.Echo, config.MaxRoomCache),
		mtx:   &sync.Mutex{},
		users: make(map[string]chan *protoc.Echo, config.MaxRoomCache),
		cache: qlist.New(),
	}
}

// Pub 用户发布消息
func (r *Room) Pub(msg *protoc.Echo) {
	// @todo update populars
	r.mtx.Lock()
	defer r.mtx.Unlock()
	for _, ch := range r.users {
		ch <- msg
	}
	r.cache.PushBack(msg)
	for r.cache.Len() > 50 {
		r.cache.PopFront()
	}
}

// Entry 用户进入房间
func (r *Room) Entry(username string) ([]*protoc.Echo, <-chan *protoc.Echo) {
	ch := make(chan *protoc.Echo, config.MaxRoomCache)
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.users[username] = ch
	cache := r.cache.Elements()
	histories := make([]*protoc.Echo, 0)
	for _, msg := range cache {
		histories = append(histories, msg.(*protoc.Echo))
	}
	log.Printf("%s entered room %d", username, r.Id)
	return histories, ch
}

// Leave 用户离开
func (r *Room) Leave(username string) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	log.Printf("%s leaved room %d", username, r.Id)
	delete(r.users, username)
}

// UserCount 房间用户数
func (r *Room) UserCount() int {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	return len(r.users)
}

// PopularWords 房间频率最高的词
func (r *Room) PopularWords() []string { return nil }

func (r *Room) String() string {
	return fmt.Sprintf("\n房间号: %-3d 当前在线用户: %d\n", r.Id, r.UserCount())
}
