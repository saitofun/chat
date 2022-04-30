package users

import (
	"sync"
	"time"

	"github.com/saitofun/chat/pkg/errors"
	"github.com/saitofun/chat/pkg/models"
	"github.com/saitofun/qlib/net/qsock"
)

type mgr struct {
	users   map[string]*models.User
	clients map[string]*models.UserInfo
	mtx     *sync.Mutex
}

var c = &mgr{
	users:   make(map[string]*models.User),
	clients: make(map[string]*models.UserInfo),
	mtx:     &sync.Mutex{},
}

func Controller() *mgr { return c }

func (m *mgr) GetByName(name string) *models.User {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	return m.users[name]
}

func (m *mgr) GetByClientID(cid string) *models.UserInfo {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	return m.clients[cid]
}

func (m *mgr) CreateUser(name string, c *qsock.Node) (*models.UserInfo, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	if _, ok := m.users[name]; ok {
		return nil, errors.ErrUserExisted
	}
	user := &models.User{Name: name, CreatedAt: time.Now()}
	info := models.NewUserInfo(user, c)
	m.clients[c.ID()] = info
	m.users[name] = user
	return info, nil
}

func (m *mgr) UserLogin(name string, c *qsock.Node) (*models.UserInfo, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	var (
		user *models.User
		info *models.UserInfo
		ok   bool
	)
	if user, ok = m.users[name]; !ok {
		return nil, errors.ErrUserNotExisted
	}
	if info, ok = m.clients[c.ID()]; ok {
		return nil, errors.ErrUserOnline
	}
	info = models.NewUserInfo(user, c)
	m.clients[c.ID()] = info
	return info, nil
}

func (m *mgr) UserOffline(name string, cid string) {

}