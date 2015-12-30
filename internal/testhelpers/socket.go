package testhelpers

import (
	"github.com/gorilla/websocket"
)

//Create lobby with id 1
func SocketCreateLobby(conn *websocket.Conn) {
	args := map[string]interface{}{
		"id": "1",
		"data": map[string]interface{}{
			"request":        "lobbyCreate",
			"map":            "cp_badlands",
			"type":           "6s",
			"league":         "etf2l",
			"server":         "testerino",
			"rconpwd":        "testerino",
			"whitelistID":    "123",
			"mumbleRequired": true,
		}}
	conn.WriteJSON(args)

	ReadMessages(conn, 2, nil)
}

//Join lobby
func SocketJoinLobby(conn *websocket.Conn) {
	args := map[string]interface{}{
		"id": "1",
		"data": map[string]interface{}{
			"request": "lobbySpectatorJoin",
			"id":      1,
		},
	}
	conn.WriteJSON(args)
	ReadMessages(conn, 2, nil)

	args = map[string]interface{}{
		"id": "1",
		"data": map[string]interface{}{
			"request": "lobbyJoin",
			"id":      1,
			"team":    "blu",
			"class":   "scout1",
		},
	}
	conn.WriteJSON(args)
	ReadMessages(conn, 4, nil)

}
