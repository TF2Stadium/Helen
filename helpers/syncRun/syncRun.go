// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package helpers

import (
	"sync"

	"github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/models"
)

type collAndId struct {
	id   uint
	coll string
}

var mutexStore = make(map[collAndId]*sync.Mutex)

type arbFunc func(interface{})

func SyncRunOn(id uint, typeName string, obj interface{}, fn arbFunc) error {
	key := collAndId{id, typeName}
	mutex, ok := mutexStore[key]
	if !ok {
		mutex = &sync.Mutex{}
		mutexStore[key] = mutex
	}

	mutex.Lock()

	err := database.DB.First(obj, id).Error
	if err != nil {
		return err
	}

	fn(obj)
	mutex.Unlock()

	return nil
}

type lobbyFunc func(*models.Lobby)

func SyncRunOnLobby(id uint, fn lobbyFunc) error {
	holder := &models.Lobby{}
	return SyncRunOn(id, "lobbies", holder, func(param interface{}) {
		lobb := param.(*models.Lobby)
		fn(lobb)
	})
}
