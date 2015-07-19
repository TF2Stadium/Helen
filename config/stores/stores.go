package stores

import (
	"github.com/TF2Stadium/Server/config"
	"github.com/TF2Stadium/Server/database"
	"github.com/antonlindstrom/pgstore"
	"github.com/gorilla/sessions"
)

// var CookieStore = sessions.NewCookieStore([]byte(Constants.SessionName))
var SessionStore sessions.Store

var SocketAuthStore = make(map[string]*sessions.Session)

func SetupStores() {
	SessionStore = pgstore.NewPGStore(database.DbUrl, []byte(config.Constants.SessionName))
	SessionStore.(*pgstore.PGStore).Options.HttpOnly = true
}
