// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllers

import (
	"html/template"
	"net/http"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/controllerhelpers"
)

var (
	playerTempl = template.Must(template.New("player").Parse(
		`
<html>
  <body>
    <a href="/logout">Logout here</a> <br>
    Logged in as {{.Name}} ({{.SteamID}}) on Steam <br>
    {{if .TwitchName}} Twitch connect connected - <a href="http://twitch.tv/{{.TwitchName}}">{{.TwitchName}}</a> {{end}} <br>
  </body>
</html>
`))
	login = template.Must(template.New("login").Parse(
		`
<html>
<body>
<a href='/startLogin'>Login</a> <br>
{{if .mocklogin}}
<form action="/startMockLogin" method="get">
  <fieldset>
    <legend>Mock Login:</legend>
    SteamID:<br>
    <input type="text" name="steamid"><br>
    <input type="submit" value="Submit">
  </fieldset>
</form>
{{end}}
</body>
</html>
`))
)

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
		login.Execute(w, map[string]bool{
			"mocklogin": config.Constants.MockupAuth,
		})
	}
}
