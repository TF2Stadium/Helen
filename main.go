// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package main

import (
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/tylerb/graceful.v1"

	"github.com/DSchalla/go-pid"
	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/config/stores"
	"github.com/TF2Stadium/Helen/controllers"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/controllers/socket"
	"github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/database/migrations"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/helpers/authority"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/Helen/rpc"
	"github.com/TF2Stadium/wsevent"
	"github.com/gorilla/context"
	_ "github.com/rakyll/gom/http"
	"github.com/rs/cors"
)

func main() {
	helpers.InitLogger()
	config.SetupConstants()
	go rpc.StartRPC()

	if config.Constants.ProfilerEnable {
		address := "localhost:" + config.Constants.ProfilerPort
		go func() {
			graceful.Run(address, 1*time.Second, nil)
		}()
		helpers.Logger.Info("Running Profiler at %s", address)
	}

	pid := &pid.Instance{}
	if pid.Create() == nil {
		defer pid.Remove()
	}

	authority.RegisterTypes()
	helpers.InitAuthorization()
	database.Init()
	migrations.Do()
	stores.SetupStores()
	models.FumbleConnect()
	models.InitializeLobbySettings("./lobbySettingsData.json")

	if !config.Constants.ServerMockUp {
		models.CheckConnection()
	}

	chelpers.InitGeoIPDB()
	if config.Constants.SteamIDWhitelist != "" {
		go chelpers.WhitelistListener()
	}
	// lobby := models.NewLobby("cp_badlands", 10, "a", "a", 1)

	// init http server

	// init socket.io server
	server := wsevent.NewServer()
	nologin := wsevent.NewServer()

	broadcaster.Init(server, nologin)
	socket.ServerInit(server, nologin)
	mux := http.NewServeMux()
	controllers.SetupHTTPRoutes(mux, server, nologin)

	if val := os.Getenv("DEPLOYMENT_ENV"); strings.ToLower(val) != "production" {
		// init static FileServer
		// TODO be careful to set this to correct location when deploying
		mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, r.URL.Path[1:])
		})
	}
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   config.Constants.AllowedCorsOrigins,
		AllowCredentials: true,
	}).Handler(context.ClearHandler(mux))

	// start the server
	helpers.Logger.Info("Serving at %s", config.Constants.Domain)
	graceful.Run(":"+config.Constants.Port, 1*time.Second, corsHandler)
}
