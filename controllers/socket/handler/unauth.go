package handler

import (
	"fmt"

	"github.com/TF2Stadium/Helen/controllers/controllerhelpers/hooks"
	"github.com/TF2Stadium/Helen/controllers/socket/sessions"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/Helen/routes/socket"
	"github.com/TF2Stadium/wsevent"
)

type Unauth struct{}

func (Unauth) Name(s string) string {
	return string((s[0])+32) + s[1:]
}

func (Unauth) LobbySpectatorJoin(so *wsevent.Client, args struct {
	ID *uint `json:"id"`
}) interface{} {

	var lob *models.Lobby
	lob, tperr := models.GetLobbyByID(*args.ID)

	if tperr != nil {
		return tperr
	}

	hooks.AfterLobbySpec(socket.UnauthServer, so, lob)

	so.EmitJSON(helpers.NewRequest("lobbyData", models.DecorateLobbyData(lob, true)))

	return emptySuccess
}

func (Unauth) LobbySpectatorLeave(so *wsevent.Client, args struct {
	ID *uint `json:"id"`
}) interface{} {

	id, ok := sessions.GetSpectating(so.ID)
	if ok {
		socket.UnauthServer.Leave(so, fmt.Sprintf("%d_public", id))
		sessions.RemoveSpectator(so.ID)
	}

	return emptySuccess
}

func (Unauth) PlayerProfile(so *wsevent.Client, args struct {
	Steamid *string `json:"steamid"`
}) interface{} {

	player, err := models.GetPlayerBySteamID(*args.Steamid)
	if err != nil {
		return err
	}

	player.SetPlayerProfile()
	return newResponse(player)
}
