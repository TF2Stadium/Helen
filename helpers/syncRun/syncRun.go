package helpers

import (
	"sync"

	"github.com/TF2Stadium/Server/database"
	"github.com/TF2Stadium/Server/models/lobby"

	"gopkg.in/mgo.v2/bson"
)

type collAndId struct {
	id   bson.ObjectId
	coll string
}

var mutexStore = make(map[collAndId]*sync.Mutex)

type arbFunc func(interface{})

func SyncRunOn(id bson.ObjectId, collection string, obj interface{}, fn arbFunc) error {
	key := collAndId{id, collection}
	mutex, ok := mutexStore[key]
	if !ok {
		mutex = &sync.Mutex{}
		mutexStore[key] = mutex
	}

	mutex.Lock()

	err := database.GetCollection(collection).FindId(id).One(obj)
	if err != nil {
		return err
	}

	fn(obj)
	mutex.Unlock()

	return nil
}

type lobbyFunc func(*lobby.Lobby)

func SyncRunOnLobby(id bson.ObjectId, fn lobbyFunc) error {
	holder := &lobby.Lobby{}
	return SyncRunOn(id, "lobbies", holder, func(param interface{}) {
		lobb := param.(*lobby.Lobby)
		fn(lobb)
	})
}
