package admin

import (
	"html/template"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/models/lobby"
)

var lobbiesTempl *template.Template

func ViewOpenLobbies(w http.ResponseWriter, r *http.Request) {
	var lobbies []*lobby.Lobby
	db.DB.Model(&lobby.Lobby{}).Preload("ServerInfo").Where("state = ?", lobby.InProgress).Find(&lobbies)
	err := lobbiesTempl.Execute(w, map[string]interface{}{
		"Lobbies":     lobbies,
		"FrontendURL": config.Constants.LoginRedirectPath,
	})
	if err != nil {
		logrus.Error(err)
	}
}
