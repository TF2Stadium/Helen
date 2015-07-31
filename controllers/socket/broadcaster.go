package socket

import (
	"time"

	db "github.com/TF2Stadium/Server/database"
	"github.com/TF2Stadium/Server/decorators"
	"github.com/TF2Stadium/Server/helpers"
	"github.com/TF2Stadium/Server/models"
	"github.com/googollee/go-socket.io"
)

var broadcasterTicker *time.Ticker
var broadcastStopChannel chan bool
var socketServer *socketio.Server

func InitBroadcaster(server *socketio.Server) {
	broadcasterTicker = time.NewTicker(time.Millisecond * 500)
	broadcastStopChannel = make(chan bool)
	socketServer = server
	go broadcaster()
}

func StopBroadcaster() {
	broadcasterTicker.Stop()
	broadcastStopChannel <- true
}

func broadcaster() {
	for {
		select {
		case <-broadcasterTicker.C:
			var lobbies []models.Lobby
			db.DB.Where("state = ?", models.LobbyStateWaiting).Order("id desc").Find(&lobbies)
			list, err := decorators.GetLobbyListData(lobbies)
			if err != nil {
				helpers.Logger.Warning("Failed to send lobby list: %s", err.Error())
			} else {
				socketServer.BroadcastTo("-1", "lobbyListData", list)
			}

		case <-broadcastStopChannel:
			return
		}
	}
}
