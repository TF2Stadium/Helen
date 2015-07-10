package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"gopkg.in/tylerb/graceful.v1"

	"github.com/TeamPlayTF/Server/config"
	"github.com/TeamPlayTF/Server/database"
	"github.com/TeamPlayTF/Server/routes"
	"github.com/googollee/go-socket.io"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	config.SetupConstants()
	database.Init()

	// lobby := models.NewLobby("cp_badlands", 10, "a", "a", 1)
	fmt.Println("Starting the server")

	r := mux.NewRouter()

	// init http server
	routes.SetupHTTPRoutes(r)
	http.Handle("/", r)

	// init socket.io server
	socketServer, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
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
	log.Println("Serving at localhost:" + config.Constants.Port + "...")
	graceful.Run(":"+config.Constants.Port, 10*time.Second, corsHandler)
}
