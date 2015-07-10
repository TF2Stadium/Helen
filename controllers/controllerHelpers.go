package controllers

import (
	"net/http"

	"github.com/TeamPlayTF/Server/config"
	"github.com/gorilla/sessions"
)

//Response structure
//github.com/TeamPlayTF/Specifications/blob/master/Communication.md#response-format
type Response struct {
	Successful bool        `json:"successful"` //true if operation was successful
	Data       interface{} `json:"data"`       //response message, if any
	Code       int         `json: "code"`      //errcode, if sucessful == false
}

func SendJSON(w http.ResponseWriter, j string) {
	w.Header().Add("Content-Type", "application/json")
	fmt.Fprintf(w, j)
}

func SendError(w http.ResponseWriter, code int, message string) string {
	r := &Response{
		Successful: false,
		Data:       data,
		Code:       code,
	}
	j, _ := json.Marshall(r)
	sendJSON(w, j)
	return string(j)
}

func SendSuccess(w http.ResponseWriter, data interface{}) string {
	r := &Response{
		Successful: true,
		Data:       data,
		Code:       -1,
	}
	j, _ := json.Marshall(r)
	sendJSON(w, j)
	return string(j)
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
