package controllers

import (
	"fmt"
	"net/http"

	"github.com/TeamPlayTF/Server/config"
	"github.com/bitly/go-simplejson"
	"github.com/gorilla/sessions"
)

func sendJSON(w http.ResponseWriter, json *simplejson.Json) {
	w.Header().Add("Content-Type", "application/json")
	val, _ := json.String()
	fmt.Fprintf(w, val)
}

func buildSuccessJSON(data simplejson.Json) *simplejson.Json {
	j := simplejson.New()
	j.Set("success", true)
	j.Set("data", data)

	return j
}

func buildFailureJSON(code int, message string) *simplejson.Json {
	j := simplejson.New()
	j.Set("success", false)
	j.Set("message", message)
	j.Set("code", code)

	return j
}

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
