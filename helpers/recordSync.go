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

func LockRecord(id uint, recType interface{}) {
	key := record{id, reflect.TypeOf(recType)}
	mutex, e := mutexStore[key]
	if !e {
		mutex = &sync.Mutex{}
		mutexStore[key] = mutex
	}

	mutex.Lock()
}

func UnlockRecord(id uint, recType interface{}) {
	key := record{id, reflect.TypeOf(recType)}
	mutex, e := mutexStore[key]
	if e {
		mutex.Unlock()
	}
}
