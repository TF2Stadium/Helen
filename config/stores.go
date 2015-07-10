package config

import "github.com/gorilla/sessions"

var CookieStore = sessions.NewCookieStore([]byte("cookie-secret"))
