package controllerhelpers

import (
	"fmt"
	"net/http"

	"github.com/TF2Stadium/Server/config"
	"github.com/TF2Stadium/Server/helpers"
	"github.com/bitly/go-simplejson"
)

func SendJSON(w http.ResponseWriter, json *simplejson.Json) {
	w.Header().Add("Content-Type", "application/json")
	val, _ := json.String()
	fmt.Fprintf(w, val)
}

func BuildSuccessJSON(data *simplejson.Json) *simplejson.Json {
	j := simplejson.New()
	j.Set("success", true)
	j.Set("data", data)

	return j
}

func BuildFailureJSON(message string, code int) *simplejson.Json {
	e := helpers.NewTPError(message, code)
	return e.ErrorJSON()
}

func RedirectHome(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, config.Constants.Domain, 303)
}
