package config

import (
	"github.com/gorilla/sessions"
)

var CookieStore = sessions.NewCookieStore([]byte(Constants.SessionName))

func SetupStores() {
	CookieStore.Options.HttpOnly = false
}
