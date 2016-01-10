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

var Fumble *rpc.Client

var FumbleLobbiesLock = new(sync.RWMutex)
var FumbleLobbies = make(map[uint]*mumble.Lobby)

func FumbleConnect() {
	if config.Constants.FumblePort == "" {
		return
	}

	helpers.Logger.Debug("Connecting to Fumble on port %s", config.Constants.FumblePort)
	client, err := rpc.DialHTTP("tcp", "localhost:"+config.Constants.FumblePort)
	if err != nil {
		helpers.Logger.Fatal(err)
	}

	Fumble = client
	helpers.Logger.Debug("Connected!")
}

func FumbleReconnect() {
	if config.Constants.FumblePort == "" {
		return
	}

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

func FumbleLobbyCreated(lob *Lobby) error {
	if Fumble == nil {
		return nil
	}

	err := Fumble.Call("Fumble.CreateLobby", lob.ID, &struct{}{})

	if err != nil {
		helpers.Logger.Warning(err.Error())
		return err
	}
	FumbleLobbiesLock.Lock()
	defer FumbleLobbiesLock.Unlock()

	return nil
}

func fumbleAllowPlayer(lobbyId uint, playerName string, playerTeam string) error {
	if Fumble == nil {
		return nil
	}

	user := mumble.User{}
	user.Name = playerName
	user.Team = mumble.Team(playerTeam)

	err := Fumble.Call("Fumble.AllowPlayer", &mumble.LobbyArgs{user, lobbyId}, &struct{}{})
	if err != nil {
		helpers.Logger.Warning(err.Error())
	}

	return nil
}

func FumbleLobbyStarted(lob_ *Lobby) {
	if Fumble == nil {
		return
	}

	var lob Lobby
	db.DB.Preload("Slots").First(&lob, lob_.ID)

	for _, slot := range lob.Slots {
		team, class, _ := LobbyGetSlotInfoString(lob.Type, slot.Slot)

		var player Player
		db.DB.First(&player, slot.PlayerID)

		if _, ok := broadcaster.GetSocket(player.SteamID); ok {
			/*var userIp string
			if userIpParts := strings.Split(so.Request().RemoteAddr, ":"); len(userIpParts) == 2 {
				userIp = userIpParts[0]
			} else {
				userIp = so.Request().RemoteAddr
			}*/
			fumbleAllowPlayer(lob.ID, strings.ToUpper(class)+"_"+player.Name, strings.ToUpper(team))
		}
	}
}

func FumbleLobbyPlayerJoinedSub(lob *Lobby, player *Player, slot int) {
	if Fumble == nil {
		// TODO fix
		return
	}

	team, class, _ := LobbyGetSlotInfoString(lob.Type, slot)
	fumbleAllowPlayer(lob.ID, strings.ToUpper(class)+"_"+player.Name, strings.ToUpper(team))
}

func FumbleLobbyPlayerJoined(lob *Lobby, player *Player, slot int) {
	if Fumble == nil {
		// TODO fix
		return
	}

	_, class, _ := LobbyGetSlotInfoString(lob.Type, slot)
	fumbleAllowPlayer(lob.ID, strings.ToUpper(class)+"_"+player.Name, "")
}

func FumbleLobbyEnded(lob *Lobby) {
	if Fumble == nil {
		return
	}

	FumbleLobbiesLock.Lock()
	defer FumbleLobbiesLock.Unlock()

	err := Fumble.Call("Fumble.EndLobby", lob.ID, nil)
	if err != nil {
		helpers.Logger.Warning(err.Error())
	}
	delete(FumbleLobbies, lob.ID)
}
