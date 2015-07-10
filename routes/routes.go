package routes

import (
	"github.com/TeamPlayTF/Server/controllers"
	"github.com/googollee/go-socket.io"
	"github.com/gorilla/mux"
)

func SetupHTTPRoutes(router *mux.Router) {
	router.HandleFunc("/", controllers.MainHandler)
	router.HandleFunc("/openidcallback", controllers.LoginCallbackHandler)
	router.HandleFunc("/startLogin", controllers.LoginHandler)
	router.HandleFunc("/{param}", controllers.ExampleHandler)

}

func SetupSocketRoutes(server *socketio.Server) {
	server.On("connection", controllers.SocketInit)
}
