package controllers

import (
	"github.com/TF2Stadium/Helen/controllers/admin"
	"github.com/TF2Stadium/Helen/models"
	"html/template"
	"net/http"
	"regexp"
)

var (
	twitchBadge *template.Template
	reValidPath = regexp.MustCompile(`badge/(\d+)`)
)

func TwitchBadge(w http.ResponseWriter, r *http.Request) {
	if !reValidPath.MatchString(r.URL.Path) {
		http.NotFound(w, r)
		return
	}

	matches := reValidPath.FindStringSubmatch(r.URL.Path)
	steamid := matches[1]

	player, err := models.GetPlayerBySteamID(steamid)
	if err != nil { //player not found
		http.Error(w, "Player with given SteamID not found", http.StatusNotFound)
		return
	}

	id, err := player.GetLobbyID(false)
	if err != nil {
		//player not in lobby right now
		return
	}

	lobby, _ := models.GetLobbyByID(id)
	err = twitchBadge.Execute(w, lobby)
}

func InitTemplates() {
	admin.InitAdminTemplates()
	twitchBadge = template.Must(template.ParseFiles("views/twitchbadge.html"))
}
