package lobby

import (
	"sync"

	db "github.com/TF2Stadium/Helen/database"
)

var (
	mu         = new(sync.RWMutex)
	lobbyLocks = make(map[uint]*sync.Mutex)
)

//Lock aquires the lock for the given lobby.
//Be careful while using Lock outside of models,
//improper usage could result in deadlocks
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
	lobbyLocks[lobby.ID] = new(sync.Mutex)
	mu.Unlock()
}

func (lobby *Lobby) deleteLock() {
	mu.Lock()
	lock, ok := lobbyLocks[lobby.ID]
	if ok {
		lock.Lock()
		delete(lobbyLocks, lobby.ID)
		lock.Unlock()
	}
	mu.Unlock()
}

//CreateLocks creates locks for all lobbies that haven't ended yet
func CreateLocks() {
	mu.Lock()
	defer mu.Unlock()

	var ids []uint

	db.DB.Model(&Lobby{}).Where("state <> ?", Ended).Pluck("id", &ids)
	for _, id := range ids {
		lobbyLocks[id] = new(sync.Mutex)
	}
}
