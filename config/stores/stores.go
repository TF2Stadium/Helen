package stores

import (
	"github.com/TF2Stadium/Server/config"
	"github.com/TF2Stadium/Server/database"
	"github.com/gorilla/sessions"
	"gopkg.in/bluesuncorp/mongostore.v4"
)

// var CookieStore = sessions.NewCookieStore([]byte(Constants.SessionName))
var SessionStore sessions.Store

var SocketAuthStore = make(map[string]*sessions.Session)

func SetupStores() {
	// get the collection in a new mongodb connection
	dbSession, _ := database.Get("sessions")
	SessionStore = mongostore.New(dbSession, "sessions", &sessions.Options{HttpOnly: false}, true, []byte(config.Constants.SessionName))
}
