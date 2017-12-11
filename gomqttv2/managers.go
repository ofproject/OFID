package gomqttv2

import (
	"sync"
)

type Manager struct {
	connects map[string]*Handler
	lock     sync.Mutex
}

var (
	manager *Manager = nil
)

func InitManagers() {
	if manager == nil {
		manager = new(Manager)
		manager.connects = make(map[string]*Handler)
	}
}

func (m *Manager) checkHander(clientid string) *Handler {
	v, ok := m.connects[clientid]
	if ok {
		return v
	}

	return nil
}

func (m *Manager) delHander(clintid string) {
	m.lock.Lock()
	delete(m.connects, clintid)
	m.lock.Unlock()
}

func (m *Manager) counts() int {
	return len(m.connects)
}

func (m *Manager) insterHander(h *Handler) bool {
	m.lock.Lock()

	_, ok := m.connects[h.cliendID]
	if ok {
		m.lock.Unlock()
		return false
	}

	m.connects[h.cliendID] = h
	m.lock.Unlock()
	return true
}
