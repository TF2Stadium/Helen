package models

import (
	"net/rpc"
	"strings"
	"sync"
	"time"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/fumble/mumble"
)

var FumbleLock = new(sync.RWMutex)
var Fumble *rpc.Client

func FumbleConnect() {
	FumbleLock.Lock()
	defer FumbleLock.Unlock()

	helpers.Logger.Debug("Connecting to Fumble on port %s", config.Constants.FumblePort)
	client, err := rpc.DialHTTP("tcp", "localhost:"+config.Constants.FumblePort)
	if err != nil {
		helpers.Logger.Fatal(err)
	}

	Fumble = client
	helpers.Logger.Debug("Connected!")
}

func FumbleReconnect() {
	FumbleLock.Lock()
	defer FumbleLock.Unlock()

	helpers.Logger.Debug("Reconnecting to Fumble on port %s", config.Constants.FumblePort)
	client, err := rpc.DialHTTP("tcp", "localhost:"+config.Constants.FumblePort)

	for err != nil {
		helpers.Logger.Critical("%s", err.Error())
		time.Sleep(1 * time.Second)
		client, err = rpc.DialHTTP("tcp", "localhost:"+config.Constants.FumblePort)
	}

	Fumble = client
	helpers.Logger.Debug("Connected!")
}

func FumbleLobbyCreated(lob *Lobby) *mumble.Lobby {
	newLobby := mumble.NewLobby()
	newLobby.ID = int(lob.ID)

	lobby := new(mumble.Lobby)

	err := Fumble.Call("Fumble.CreateLobby", &newLobby, &lobby)

	if err != nil {
		helpers.Logger.Warning(err.Error())
		return nil
	}
	return lobby
}

func FumbleLobbyStarted(lob_ *Lobby) {
	var lob Lobby
	db.DB.Preload("Slots").First(&lob, lob_.ID)

	for _, slot := range lob.Slots {
		team, class, _ := LobbyGetSlotInfoString(lob.Type, slot.Slot)

		var player Player
		db.DB.First(&player, slot.PlayerId)

		if _, ok := broadcaster.GetSocket(player.SteamId); ok {
			/*var userIp string
			if userIpParts := strings.Split(so.Request().RemoteAddr, ":"); len(userIpParts) == 2 {
				userIp = userIpParts[0]
			} else {
				userIp = so.Request().RemoteAddr
			}*/

			user := mumble.NewUser()
			user.Name = strings.ToUpper(class) + " " + player.Name
			lobby := mumble.NewLobby()
			lobby.ID = int(lob.ID)

			Fumble.Call("Fumble.AddNameToLobbyWhitelist", mumble.LobbyArgsTeam{user, lobby, strings.ToUpper(team)}, nil)

			/*user := mumble.NewUser()
			Fumble.Call("Fumble.FindUserByIP", userIp, &user)

			if user.UserID != 0 {
				Fumble.Call("Fumble.AllowPlayer", &mumble.LobbyArgs{}, nil)
			}*/
		} else {
			helpers.Logger.Warning("Socket for player with steamid[%d] not found.", player.SteamId)
		}
	}
}

func FumbleLobbyPlayerJoinedSub(lob *Lobby, player *Player, slot int) {
	team, class, _ := LobbyGetSlotInfoString(lob.Type, slot)

	user := mumble.NewUser()
	user.Name = strings.ToUpper(class) + " " + player.Name
	lobby := mumble.NewLobby()
	lobby.ID = int(lob.ID)
	Fumble.Call("Fumble.AddNameToLobbyWhitelist", mumble.LobbyArgsTeam{user, lobby, strings.ToUpper(team)}, nil)
}

func FumbleLobbyPlayerJoined(lob *Lobby, player *Player, slot int) {
	_, class, _ := LobbyGetSlotInfoString(lob.Type, slot)

	user := mumble.NewUser()
	user.Name = strings.ToUpper(class) + " " + player.Name
	lobby := mumble.NewLobby()
	lobby.ID = int(lob.ID)
	Fumble.Call("Fumble.AddNameToLobbyWhitelist", mumble.LobbyArgsTeam{user, lobby, ""}, nil)
}

func FumbleLobbyEnded(lobby *Lobby) {

}
