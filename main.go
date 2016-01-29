// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package main

import (
	"encoding/base64"
	_ "expvar"
	"flag"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"
	"time"

	"gopkg.in/tylerb/graceful.v1"

	"github.com/DSchalla/go-pid"
	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/config/stores"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/controllers/socket"
	"github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/database/migrations"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/helpers/authority"
	_ "github.com/TF2Stadium/Helen/internal/pprof"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/Helen/routes"
	"github.com/TF2Stadium/Helen/rpc"
	"github.com/gorilla/context"
	"github.com/gorilla/securecookie"
	"github.com/rs/cors"
)

var flagGen = flag.String("genkey", "", "write a 32bit key for encrypting cookies the given file, and exit")

func main() {
	helpers.InitLogger()

	flag.Parse()
	if *flagGen != "" {
		key := securecookie.GenerateRandomKey(64)
		if len(key) == 0 {
			logrus.Fatal("Couldn't generate random key")
		}

		base64Key := base64.StdEncoding.EncodeToString(key)
		ioutil.WriteFile(*flagGen, []byte(base64Key), 0666)
		logrus.Infof("Written key to file %s", *flagGen)
		return
	}

	config.SetupConstants()
	go rpc.StartRPC()

	if config.Constants.ProfilerEnable {
		address := "localhost:" + config.Constants.ProfilerPort
		go func() {
			graceful.Run(address, 1*time.Second, nil)
		}()
		logrus.Info("Running Profiler at %s", address)
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
	models.InitializeLobbySettings("./lobbySettingsData.json")

	models.ConnectRPC()
	models.DeleteUnusedServerRecords()
	go models.Ping()

	chelpers.InitGeoIPDB()
	if config.Constants.SteamIDWhitelist != "" {
		go chelpers.WhitelistListener()
	}
	// lobby := models.NewLobby("cp_badlands", 10, "a", "a", 1)

	mux := http.NewServeMux()
	routes.SetupHTTP(mux)
	socket.RegisterHandlers()

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
	logrus.Info("Serving at %s", config.Constants.Domain)
	graceful.Run(":"+config.Constants.Port, 1*time.Second, corsHandler)
}
