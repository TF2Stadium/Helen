// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllers

import (
	"fmt"
	"net/http"

	"github.com/TF2Stadium/Helen/controllers/controllerhelpers"
)

func MainHandler(w http.ResponseWriter, r *http.Request) {
	if controllerhelpers.IsLoggedInHTTP(r) {
		session, _ := controllerhelpers.GetSessionHTTP(r)
		var steamid = session.Values["steam_id"].(string)
		fmt.Fprintf(w, `<html><head></head><body>hello! You're logged in and your steam id is
			`+steamid+`. You can log out <a href='/logout'>here</a></body></html>`)
	} else {
		fmt.Fprintf(w, "<html><head></head><body>hello! You can log in <a href='/startLogin'>here</a></body></html>")
	}
}
