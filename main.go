package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"gopkg.in/tylerb/graceful.v1"

	"os"

	"github.com/TeamPlayTF/Server/routes"
	"github.com/googollee/go-socket.io"
	"github.com/gorilla/mux"
)

func main() {
	// lobby := models.NewLobby("cp_badlands", 10, "a", "a", 1)
	fmt.Print("Starting the server")

	port := ":" + os.Getenv("PORT")
	if port == ":" {
		port += "8080"
	}

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
		http.FileServer(http.Dir(os.Getenv("GOPATH")+"/src/github.com/TeamPlayTF/Server/static"))))

	// start the server
	graceful.Run(port, 10*time.Second, r)

	log.Println("Serving at localhost:" + port + "...")
}
