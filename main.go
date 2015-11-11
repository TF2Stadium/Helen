// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package main

import (
	"net/http"
	"time"

	"gopkg.in/tylerb/graceful.v1"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/config/stores"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/controllers/socket"
	"github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/database/migrations"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/helpers/authority"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/Helen/routes"
	"github.com/TF2Stadium/wsevent"
	"github.com/DSchalla/go-pid"
	"github.com/rs/cors"
)

func main() {

	pid := &pid.Instance {}
	if pid.Create() == nil {
		defer pid.Remove()
	}

	authority.RegisterTypes()
	helpers.InitLogger()
	helpers.InitAuthorization()
	config.SetupConstants()
	database.Init()
	migrations.Do()
	stores.SetupStores()
	models.PaulingConnect()
	models.InitializeLobbySettings("./lobbySettingsData.json")

	go models.ReadyTimeoutListener()
	StartPaulingListener()
	chelpers.CheckLogger()
	if config.Constants.SteamIDWhitelist != "" {
		go chelpers.WhitelistListener()
	}
	// lobby := models.NewLobby("cp_badlands", 10, "a", "a", 1)
	helpers.Logger.Debug("Starting the server")

	// init http server

	// init socket.io server
	server := wsevent.NewServer()
	broadcaster.Init(server)
	socket.ServerInit(server)
	routes.SetupHTTPRoutes(server)
	go server.Listener()

	// init static FileServer
	// TODO be careful to set this to correct location when deploying
	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path[1:])
	})
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   config.Constants.AllowedCorsOrigins,
		AllowCredentials: true,
	}).Handler(http.DefaultServeMux)

	// start the server
	helpers.Logger.Debug("Serving at %s", config.Constants.Domain)
	graceful.Run(":"+config.Constants.Port, 10*time.Second, corsHandler)
}
