package main

import (
	"net/http"
	"time"

	"gopkg.in/tylerb/graceful.v1"

	"github.com/TF2Stadium/Server/config"
	"github.com/TF2Stadium/Server/config/stores"
	"github.com/TF2Stadium/Server/database"
	"github.com/TF2Stadium/Server/database/migrations"
	"github.com/TF2Stadium/Server/helpers"
	"github.com/TF2Stadium/Server/models"
	"github.com/TF2Stadium/Server/routes"
	"github.com/googollee/go-socket.io"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	config.SetupConstants()
	helpers.InitLogger()
	database.Init()
	migrations.Do()
	stores.SetupStores()
	models.InitServerConfigs()

	// lobby := models.NewLobby("cp_badlands", 10, "a", "a", 1)
	helpers.Logger.Debug("Starting the server")

	r := mux.NewRouter()

	// init http server
	routes.SetupHTTPRoutes(r)
	http.Handle("/", r)

	// init socket.io server
	socketServer, err := socketio.NewServer(nil)
	if err != nil {
		helpers.Logger.Critical(err.Error())
	}
	routes.SetupSocketRoutes(socketServer)
	r.Handle("/socket.io/", socketServer)

	// init static FileServer
	// TODO be careful to set this to correct location when deploying
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/",
		http.FileServer(http.Dir(config.Constants.StaticFileLocation))))

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   config.Constants.AllowedCorsOrigins,
		AllowCredentials: true,
	}).Handler(r)

	// start the server
	helpers.Logger.Debug("Serving at localhost:" + config.Constants.Port + "...")
	graceful.Run(":"+config.Constants.Port, 10*time.Second, corsHandler)
}
