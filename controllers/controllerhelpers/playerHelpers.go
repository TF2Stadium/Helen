package controllerhelpers

import (
	"github.com/TF2Stadium/Helen/models"
	"github.com/googollee/go-socket.io"
	"strconv"
)

var BanTypeList = []string{"join", "create", "chat", "full"}

var BanTypeMap = map[string]models.PlayerBanType{
	"join":   models.PlayerBanJoin,
	"create": models.PlayerBanCreate,
	"chat":   models.PlayerBanChat,
	"full":   models.PlayerBanFull,
}

func AfterLobbyJoin(so socketio.Socket, lobby *models.Lobby, player *models.Player) {
	so.Join(GetLobbyRoom(lobby.ID))
}

func AfterLobbyLeave(so socketio.Socket, lobby *models.Lobby, player *models.Player) {
	so.Join(GetLobbyRoom(lobby.ID))

}

func GetLobbyRoom(lobbyid uint) string {
	return strconv.FormatUint(uint64(lobbyid), 10)
}
