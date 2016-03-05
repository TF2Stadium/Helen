package models

import (
	"sync"
)

var (
	mu         = new(sync.RWMutex)
	lobbyLocks = make(map[uint]*sync.RWMutex)
)

func (lobby *Lobby) Lock() {
	mu.RLock()
	lock, ok := lobbyLocks[lobby.ID]
	mu.RUnlock()
	if !ok {
		lock.Lock()
	}
}

func (lobby *Lobby) Unlock() {
	mu.RLock()
	lock, ok := lobbyLocks[lobby.ID]
	mu.RUnlock()
	if !ok {
		lock.Unlock()
	}
}

func (lobby *Lobby) createLock() {
	mu.Lock()
	lobbyLocks[lobby.ID] = new(sync.RWMutex)
	mu.Unlock()
}

func (lobby *Lobby) deleteLock() {
	mu.Lock()
	delete(lobbyLocks, lobby.ID)
	mu.Unlock()
}
