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

var mutexStore = make(map[record]*sync.RWMutex)
var storeLock = &sync.RWMutex{}

func RLockRecord(id uint, recType interface{}) {
	storeLock.Lock()
	key := record{id, reflect.TypeOf(recType)}
	mutex, e := mutexStore[key]
	if !e {
		mutex = &sync.RWMutex{}
		mutexStore[key] = mutex
	}
	storeLock.Unlock()

	mutex.RLock()
}

func RUnlockRecord(id uint, recType interface{}) {
	storeLock.RLock()
	defer storeLock.RUnlock()

	key := record{id, reflect.TypeOf(recType)}
	mutex, e := mutexStore[key]
	if e {
		mutex.RUnlock()
	}
}

func LockRecord(id uint, recType interface{}) {
	key := record{id, reflect.TypeOf(recType)}
	storeLock.RLock()
	mutex, e := mutexStore[key]
	storeLock.RUnlock()

	if !e {
		mutex = &sync.RWMutex{}

		storeLock.Lock()
		mutexStore[key] = mutex
		storeLock.Unlock()
	}

	mutex.Lock()
}

func UnlockRecord(id uint, recType interface{}) {
	key := record{id, reflect.TypeOf(recType)}
	storeLock.RLock()
	defer storeLock.RUnlock()

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
