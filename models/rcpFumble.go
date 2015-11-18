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

	newLobby := mumble.NewLobby()
	newLobby.ID = int(lob.ID)

	lobby := new(mumble.Lobby)

	err := Fumble.Call("Fumble.CreateLobby", &newLobby, &lobby)

	if err != nil {
		helpers.Logger.Warning(err.Error())
		return err
	}
	FumbleLobbiesLock.Lock()
	defer FumbleLobbiesLock.Unlock()
	FumbleLobbies[lob.ID] = lobby

	return nil
}

func FumbleAllowPlayer(lobbyId uint, playerName string, playerTeam string) error {
	if Fumble == nil {
		return nil
	}

	user := mumble.NewUser()
	user.Name = playerName
	user.Team = mumble.Team(playerTeam)

	FumbleLobbiesLock.Lock()
	defer FumbleLobbiesLock.Unlock()

	reply := new(mumble.Lobby)

	err := Fumble.Call("Fumble.AllowPlayer", &mumble.LobbyArgs{user, FumbleLobbies[lobbyId]}, reply)
	if err != nil {
		helpers.Logger.Warning(err.Error())
	}
	FumbleLobbiesLock.Lock()
	defer FumbleLobbiesLock.Unlock()
	FumbleLobbies[lobbyId] = reply
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
		db.DB.First(&player, slot.PlayerId)

		if _, ok := broadcaster.GetSocket(player.SteamId); ok {
			/*var userIp string
			if userIpParts := strings.Split(so.Request().RemoteAddr, ":"); len(userIpParts) == 2 {
				userIp = userIpParts[0]
			} else {
				userIp = so.Request().RemoteAddr
			}*/
			FumbleAllowPlayer(lob.ID, strings.ToUpper(class)+" "+player.Name, strings.ToUpper(team))
		} else {
			helpers.Logger.Warning("Socket for player with steamid[%d] not found.", player.SteamId)
		}
	}
}

func FumbleLobbyPlayerJoinedSub(lob *Lobby, player *Player, slot int) {
	if Fumble == nil {
		// TODO fix
		return
	}

	team, class, _ := LobbyGetSlotInfoString(lob.Type, slot)
	FumbleAllowPlayer(lob.ID, strings.ToUpper(class)+" "+player.Name, strings.ToUpper(team))
}

func FumbleLobbyPlayerJoined(lob *Lobby, player *Player, slot int) {
	if Fumble == nil {
		// TODO fix
		return
	}

	_, class, _ := LobbyGetSlotInfoString(lob.Type, slot)
	FumbleAllowPlayer(lob.ID, strings.ToUpper(class)+" "+player.Name, "")
}

func FumbleLobbyEnded(lob *Lobby) {
	if Fumble == nil {
		return
	}

	FumbleLobbiesLock.Lock()
	defer FumbleLobbiesLock.Unlock()

	err := Fumble.Call("Fumble.EndLobby", FumbleLobbies[lob.ID], nil)
	if err != nil {
		helpers.Logger.Warning(err.Error())
	}
	delete(FumbleLobbies, lob.ID)
}
