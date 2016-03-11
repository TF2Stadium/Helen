package admin

import (
	"html/template"
	"net/http"

	"github.com/TF2Stadium/Helen/config"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/models"
)

var lobbiesTempl *template.Template

func ViewOpenLobbies(w http.ResponseWriter, r *http.Request) {
	var lobbies []*models.Lobby
	db.DB.Model(&models.Lobby{}).Where("state = ?", models.LobbyStateInProgress).Find(&lobbies)
	lobbiesTempl.Execute(w, map[string]interface{}{
		"Lobbies":     lobbies,
		"FrontendURL": config.Constants.LoginRedirectPath,
	})
}
