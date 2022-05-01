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

func (m *mgr) CreateRoom(id int) (*models.Room, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	if id == 0 {
		for i := 1; ; i++ {
			if _, ok := m.Rooms[i]; !ok {
				id = i
			}
		}
	} else {
		if r, ok := m.Rooms[id]; ok {
			return r, nil
		}
	}
	ret := models.NewRoom(id)
	m.Rooms[id] = ret
	m.idx++
	return ret, nil
}

func (m *mgr) RoomList() (ret models.Rooms) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	for _, v := range m.Rooms {
		ret = append(ret, *v)
	}
	return ret
}
