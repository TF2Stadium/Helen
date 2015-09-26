// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package stores

import (
	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/database"
	"github.com/antonlindstrom/pgstore"
	"github.com/gorilla/sessions"
	"sync"
)

var sessionStoreMutex sync.Mutex

// var CookieStore = sessions.NewCookieStore([]byte(Constants.SessionName))
var SessionStore *pgstore.PGStore

var SocketAuthStore = make(map[string]*sessions.Session)

func SetupStores() {
	if SessionStore == nil {
		sessionStoreMutex.Lock()
		if SessionStore == nil {
			SessionStore = pgstore.NewPGStore(database.DbUrl, []byte(config.Constants.SessionName))
			SessionStore.Options.HttpOnly = true
		}
		sessionStoreMutex.Unlock()
	}
}
