// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package stores

import (
	"sync"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/database"
	"github.com/antonlindstrom/pgstore"
	"github.com/gorilla/sessions"
)

var sessionStoreMutex = &sync.Mutex{}
var authStoreMutex = &sync.RWMutex{}

// var CookieStore = sessions.NewCookieStore([]byte(Constants.SessionName))
var SessionStore *pgstore.PGStore

var socketAuthStore = make(map[string]*sessions.Session)

func SetupStores() {
	if SessionStore == nil {
		sessionStoreMutex.Lock()
		SessionStore = pgstore.NewPGStore(database.DbUrl, []byte(config.Constants.SessionName))
		SessionStore.Options.HttpOnly = true
		sessionStoreMutex.Unlock()
	}
}

func SetSocketSession(socketid string, session *sessions.Session) {
	authStoreMutex.Lock()
	socketAuthStore[socketid] = session
	authStoreMutex.Unlock()
}

func RemoveSocketSession(socketid string) {
	authStoreMutex.Lock()
	delete(socketAuthStore, socketid)
	authStoreMutex.Unlock()
}

func GetStore(socketid string) (session *sessions.Session, ok bool) {
	authStoreMutex.RLock()
	session, ok = socketAuthStore[socketid]
	authStoreMutex.RUnlock()
	return
}
