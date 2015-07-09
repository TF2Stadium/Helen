package controllers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func MainHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello!")
}

func ExampleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	param := vars["param"]
	fmt.Fprintf(w, "The url is /"+param)
}
