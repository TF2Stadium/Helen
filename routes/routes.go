package routes

import (
	"github.com/TeamPlayTF/Server/controllers"
	"github.com/googollee/go-socket.io"
	"github.com/gorilla/mux"
	"github.com/TeamPlayTF/Server/config"
)

func SetupHTTPRoutes(router *mux.Router) {
	router.HandleFunc("/", controllers.MainHandler)
	router.HandleFunc("/openidcallback", controllers.LoginCallbackHandler)
	router.HandleFunc("/startLogin", controllers.LoginHandler)
	router.HandleFunc("/logout", controllers.LogoutHandler)
	router.HandleFunc("/{param}", controllers.ExampleHandler)

}

func SetupSocketRoutes(server *socketio.Server) {

	var socketController func(socketio.Socket);

	if config.Constants.SocketMockUp {
		socketController = controllers.SocketMockUpInit;
	} else {
		socketController = controllers.SocketInit;
	}

	server.On("connection", socketController)
}
