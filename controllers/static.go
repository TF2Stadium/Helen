package controllers

import (
	"fmt"
	"net/http"

	"github.com/TeamPlayTF/Server/controllers/controllerhelpers"
	"github.com/gorilla/mux"
)

func MainHandler(w http.ResponseWriter, r *http.Request) {
	if controllerhelpers.IsLoggedInHTTP(r) {
		session, _ := controllerhelpers.GetSessionHTTP(r)
		var steamid = session.Values["steamid"].(string)
		fmt.Fprintf(w, `<html><head></head><body>hello! You're logged in and your steam id is
			`+steamid+`. You can log out <a href='/logout'>here</a></body></html>`)
	} else {
		fmt.Fprintf(w, "<html><head></head><body>hello! You can log in <a href='/startLogin'>here</a></body></html>")
	}
}

func ExampleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	param := vars["param"]
	fmt.Fprintf(w, "The url is /"+param)
}
