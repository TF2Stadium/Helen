package models

import (
	db "github.com/TF2Stadium/Helen/database"
	"sync"
)

var (
	mu         = new(sync.RWMutex)
	lobbyLocks = make(map[uint]*sync.RWMutex)
)

//Lock aquires the lock for the given lobby.
//Be careful while using Lock outside of models,
//since it could potentially result in a deadlock
//if the function call after the lock would try
//acquiring it too
func (lobby *Lobby) Lock() {
	mu.RLock()
	lock, ok := lobbyLocks[lobby.ID]
	mu.RUnlock()
	if ok {
		lock.Lock()
	}
}

//Unlock releases the lock for the given lobby
func (lobby *Lobby) Unlock() {
	mu.RLock()
	lock, ok := lobbyLocks[lobby.ID]
	mu.RUnlock()
	if ok {
		lock.Unlock()
	}
}

//CreateLock creates a lock for lobby
func (lobby *Lobby) CreateLock() {
	mu.Lock()
	lobbyLocks[lobby.ID] = new(sync.RWMutex)
	mu.Unlock()
}

func (lobby *Lobby) deleteLock() {
	mu.Lock()
	delete(lobbyLocks, lobby.ID)
	mu.Unlock()
}

//CreateLocks creates locks for all lobbies that haven't ended yet
func CreateLocks() {
	mu.Lock()
	defer mu.Unlock()

	var ids []uint

	db.DB.Model(&Lobby{}).Where("state <> ?", LobbyStateEnded).Pluck("id", &ids)
	for _, id := range ids {
		lobbyLocks[id] = new(sync.RWMutex)
	}
}
