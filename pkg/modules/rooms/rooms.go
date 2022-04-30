package rooms

import (
	"sync"

	"github.com/saitofun/chat/pkg/models"
)

type mgr struct {
	Rooms map[int]*models.Room
	mtx   *sync.Mutex
	idx   int
}

var c = &mgr{
	Rooms: make(map[int]*models.Room),
	mtx:   &sync.Mutex{},
	idx:   1,
}

func Controller() *mgr { return c }

func (m *mgr) GetByID(id int) *models.Room {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	return m.Rooms[id]
}

func (m *mgr) CreateRoom() *models.Room {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	ret := models.NewRoom(m.idx)
	m.idx++
	return ret
}

func (m *mgr) RoomList() (ret []models.Room) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	for _, v := range m.Rooms {
		ret = append(ret, *v)
	}
	return nil
}
