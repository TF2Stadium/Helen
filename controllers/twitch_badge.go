package controllers

import (
	"html/template"
	"net/http"
	"regexp"

	"github.com/TF2Stadium/Helen/models/lobby"
	"github.com/TF2Stadium/Helen/models/player"
)

var (
	twitchBadge *template.Template
	reValidPath = regexp.MustCompile(`badge/(\d+)`)
	emptyPage   = `
<html>
<head>
<meta http-equiv="refresh" content="5">
</head>
</html>
`
)

func TwitchBadge(w http.ResponseWriter, r *http.Request) {
	if !reValidPath.MatchString(r.URL.Path) {
		http.NotFound(w, r)
		return
	}

	matches := reValidPath.FindStringSubmatch(r.URL.Path)
	steamid := matches[1]

	player, err := player.GetPlayerBySteamID(steamid)
	if err != nil { //player not found
		http.Error(w, "Player with given SteamID not found", http.StatusNotFound)
		return
	}

	id, err := player.GetLobbyID(false)
	if err != nil {
		//player not in lobby right now, just serve a page that refreshes every 5 seconds
		w.Write([]byte(emptyPage))
		return
	}

	lobby, _ := lobby.GetLobbyByID(id)
	err = twitchBadge.Execute(w, lobby)
}
