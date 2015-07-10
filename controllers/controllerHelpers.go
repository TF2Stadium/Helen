package controllers

import (
	"net/http"

	"github.com/TeamPlayTF/Server/config"
	"github.com/bitly/go-simplejson"
	"github.com/gorilla/sessions"
)

//Response structure
//github.com/TeamPlayTF/Specifications/blob/master/Communication.md#response-format
type Response struct {
	Successful bool        `json:"successful"` //true if operation was successful
	Data       interface{} `json:"data"`       //response message, if any
	Code       int         `json: "code"`      //errcode, if sucessful == false
}

func SendJSON(w http.ResponseWriter, json simplejson.Json) {
	w.Header().Add("Content-Type", "application/json")
	fmt.Fprintf(w, string(json))
}

func BuildSuccessJSON(data simplejson.Json) simplejson.Json {
	j := simplejson.New()
	j.Set("success", true)
	j.Set("data", data)

	return j
}

func BuildFailureJSON(code int, message string) simplejson.Json {
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
