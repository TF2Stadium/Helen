// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package main

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"gopkg.in/tylerb/graceful.v1"

	"github.com/DSchalla/go-pid"
	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/controllers/socket"
	"github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/database/migrations"
	"github.com/TF2Stadium/Helen/helpers"
	_ "github.com/TF2Stadium/Helen/helpers/authority" // to register authority types
	_ "github.com/TF2Stadium/Helen/internal/pprof"    // to setup expvars
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/Helen/models/event"
	"github.com/TF2Stadium/Helen/routes"
	socketServer "github.com/TF2Stadium/Helen/routes/socket"
	"github.com/rs/cors"
)

var (
	flagGen  = flag.Bool("genkey", false, "write a 32bit key for encrypting cookies the given file, and exit")
	docPrint = flag.Bool("printdoc", false, "print the docs for environment variables, and exit.")
)

func main() {
	flag.Parse()

	if *flagGen {
		key := make([]byte, 64)
		_, err := rand.Read(key)
		if err != nil {
			logrus.Fatal(err)
		}

		base64Key := base64.StdEncoding.EncodeToString(key)
		fmt.Println(base64Key)
		return
	}
	if *docPrint {
		config.PrintConfigDoc()
		os.Exit(0)
	}

	controllers.InitTemplates()
	config.SetupConstants()
	helpers.SetServemeContext()
	//models.ReadServers()
	chelpers.SetupJWTSigning()

	if config.Constants.ProfilerAddr != "" {
		go graceful.Run(config.Constants.ProfilerAddr, 1*time.Second, nil)
		logrus.Info("Running Profiler at ", config.Constants.ProfilerAddr)
	}

	helpers.ConnectAMQP()
	event.StartListening()

	database.Init()
	migrations.Do()
	err := models.LoadLobbySettingsFromFile("assets/lobbySettingsData.json")
	if err != nil {
		logrus.Fatal(err)
	}

	models.CreateLocks()
	models.ConnectRPC()
	models.DeleteUnusedServerRecords()
	//go models.TFTVStreamStatusUpdater()

	helpers.InitGeoIPDB()
	if config.Constants.SteamIDWhitelist != "" {
		go chelpers.WhitelistListener()
	}

	mux := http.NewServeMux()
	routes.SetupHTTP(mux)
	socket.RegisterHandlers()

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{config.Constants.CORSWhitelist},
		AllowedMethods:   []string{"GET", "POST", "DELETE"},
		AllowCredentials: true,
	}).Handler(mux)

	pid := &pid.Instance{}
	if pid.Create() == nil {
		defer pid.Remove()
	}

	// start the server
	server := graceful.Server{
		Timeout: 10 * time.Second,
		Server: &http.Server{
			Addr:         config.Constants.ListenAddress,
			Handler:      corsHandler,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},

		ShutdownInitiated: func() {
			models.SendNotification(`Backend will be going down for a while for an update, click on "Reconnect" to reconnect to TF2Stadium`, 0)
			logrus.Info("Received SIGINT/SIGTERM")
			logrus.Info("waiting for GlobalWait")
			helpers.GlobalWait.Wait()
			logrus.Info("waiting for socket requests to complete.")
			socketServer.Wait()
			logrus.Info("closing all active websocket connections")
			socketServer.AuthServer.Close()
			logrus.Info("stopping event listener")
			event.StopListening()
		},
	}

	//start health checks
	if config.Constants.HealthChecks {
		controllers.StartHealthCheck()
	}
	logrus.Info("Serving on ", config.Constants.ListenAddress)
	err = server.ListenAndServe()
	if err != nil {
		logrus.Fatal(err)
	}
}
