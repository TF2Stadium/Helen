package controllers

import (
	"net/http"

	"github.com/TeamPlayTF/Server/config"
	"github.com/gorilla/sessions"
)

func isLoggedIn(r *http.Request) bool {
	session, _ := config.CookieStore.Get(r, config.Constants.SessionName)

	val, ok := session.Values["steamid"]
	return ok && val != ""
}

func getDefaultSession(r *http.Request) *sessions.Session {
	session, _ := config.CookieStore.Get(r, config.Constants.SessionName)
	return session
}

func redirectHome(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, config.Constants.Domain, 303)
}
