// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllers

import (
	"fmt"
	"net/http"

	"html/template"

	"github.com/TF2Stadium/Helen/controllers/controllerhelpers"
)

var playerTempl = template.Must(template.New("player").Parse(
	`
<html>
  <body>
    <a href="/logout">Logout here</a> <br>
    Logged in as {{.Name}} ({{.SteamID}}) on Steam <br>
    {{if .TwitchName}} Twitch connect connected - <a href="http://twitch.tv/{{.TwitchName}}">{{.TwitchName}}</a> {{end}} <br>
  </body>
</html>
`))

func MainHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	token, err := controllerhelpers.GetToken(r)

	if err == nil {
		player := controllerhelpers.GetPlayer(token)
		playerTempl.Execute(w, player)
	} else {
		fmt.Fprintf(w, `<html><head></head><body>hello! You can log in <a href='/startLogin'>here</a></body></html>`)
	}
}
