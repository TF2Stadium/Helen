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
	"github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/database/migrations"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/helpers/authority"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/Helen/routes"
	"github.com/googollee/go-socket.io"
	"github.com/rs/cors"
)

func main() {
	authority.RegisterTypes()
	helpers.InitLogger()
	helpers.InitAuthorization()
	config.SetupConstants()
	database.Init()
	migrations.Do()
	stores.SetupStores()
	models.PaulingConnect()
	go models.ReadyTimeoutListener()
	StartListener()
	chelpers.StartGlobalLogger()
	// lobby := models.NewLobby("cp_badlands", 10, "a", "a", 1)
	helpers.Logger.Debug("Starting the server")

	// init http server
	routes.SetupHTTPRoutes()

	// init socket.io server
	socketServer, err := socketio.NewServer(nil)
	if err != nil {
		helpers.Logger.Fatal(err.Error())
	}
	broadcaster.Init(socketServer)
	defer broadcaster.Stop()
	routes.SetupSocketRoutes(socketServer)
	http.Handle("/socket.io/", socketServer)

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
	helpers.Logger.Debug("Serving at localhost:" + config.Constants.Port + "...")
	graceful.Run(":"+config.Constants.Port, 10*time.Second, corsHandler)
}
