// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package helpers

import (
	"reflect"
	"sync"
)

type record struct {
	id      uint
	recType reflect.Type
}

var mutexStore = make(map[record]*sync.Mutex)
var storeLock = &sync.Mutex{}

func LockRecord(id uint, recType interface{}) {
	storeLock.Lock()
	key := record{id, reflect.TypeOf(recType)}
	mutex, e := mutexStore[key]
	if !e {
		mutex = &sync.Mutex{}
		mutexStore[key] = mutex
	}
	storeLock.Unlock()
	mutex.Lock()
}

func UnlockRecord(id uint, recType interface{}) {
	storeLock.Lock()
	defer storeLock.Unlock()
	key := record{id, reflect.TypeOf(recType)}
	mutex, e := mutexStore[key]
	if e {
		mutex.Unlock()
	}
}

func RemoveRecord(id uint, recType interface{}) {
	storeLock.Lock()
	defer storeLock.Unlock()
	key := record{id, reflect.TypeOf(recType)}
	delete(mutexStore, key)
}
