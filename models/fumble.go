package models

import (
	"strings"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/fumble/mumble"
)

func FumbleLobbyCreated(lob *Lobby) error {
	if config.Constants.FumblePort == "" {
		return nil
	}

	err := call(config.Constants.FumblePort, "Fumble.CreateLobby", lob.ID, &struct{}{})

	if err != nil {
		helpers.Logger.Warning(err.Error())
		return err
	}

	return nil
}

func fumbleAllowPlayer(lobbyId uint, playerName string, playerTeam string) error {
	if config.Constants.FumblePort == "" {
		return nil
	}

	user := mumble.User{}
	user.Name = playerName
	user.Team = mumble.Team(playerTeam)

	err := call(config.Constants.FumblePort, "Fumble.AllowPlayer", &mumble.LobbyArgs{user, lobbyId}, &struct{}{})
	if err != nil {
		helpers.Logger.Warning(err.Error())
	}

	return nil
}

func FumbleLobbyStarted(lob_ *Lobby) {
	if config.Constants.FumblePort == "" {
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
			fumbleAllowPlayer(lob.ID, strings.ToUpper(class)+"_"+sanitize(player.Name), strings.ToUpper(team))
		}
	}
}

func FumbleLobbyPlayerJoinedSub(lob *Lobby, player *Player, slot int) {
	if config.Constants.FumblePort == "" {
		return
	}

	team, class, _ := LobbyGetSlotInfoString(lob.Type, slot)
	fumbleAllowPlayer(lob.ID, strings.ToUpper(class)+"_"+sanitize(player.Name), strings.ToUpper(team))
}

func FumbleLobbyPlayerJoined(lob *Lobby, player *Player, slot int) {
	if config.Constants.FumblePort == "" {
		return
	}

	_, class, _ := LobbyGetSlotInfoString(lob.Type, slot)
	fumbleAllowPlayer(lob.ID, strings.ToUpper(class)+"_"+sanitize(player.Name), "")
}

func FumbleLobbyEnded(lob *Lobby) {
	if config.Constants.FumblePort == "" {
		return
	}

	err := call(config.Constants.FumblePort, "Fumble.EndLobby", lob.ID, nil)
	if err != nil {
		helpers.Logger.Warning(err.Error())
	}
}
