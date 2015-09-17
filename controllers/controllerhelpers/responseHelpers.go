// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllerhelpers

import (
	"fmt"
	"net/http"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/helpers"
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

func BuildEmptySuccessString() string {
	bytes, _ := BuildSuccessJSON(simplejson.New()).Encode()
	return string(bytes)
}

func BuildFailureJSON(message string, code int) *simplejson.Json {
	e := helpers.NewTPError(message, code)
	return e.ErrorJSON()
}

func BuildMissingArgJSON(arg string) *simplejson.Json {
	return BuildFailureJSON(fmt.Sprintf("Missing argument: '%s'", arg), 0)
}

func RedirectHome(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, config.Constants.Domain, 303)
}
