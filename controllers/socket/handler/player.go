package handler

import (
	"errors"
	"regexp"
	"sync"
	"time"

	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/controllers/controllerhelpers/hooks"
	"github.com/TF2Stadium/Helen/controllers/socket/sessions"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/TF2Stadium/wsevent"
)

type Player struct{}

func (Player) Name(s string) string {
	return string((s[0])+32) + s[1:]
}

func (Player) PlayerReady(so *wsevent.Client, _ struct{}) interface{} {
	player := chelpers.GetPlayer(so.Token)
	lobbyid, tperr := player.GetLobbyID(false)
	if tperr != nil {
		return tperr
	}

	lobby, tperr := models.GetLobbyByIDServer(lobbyid)
	if tperr != nil {
		return tperr
	}

	if lobby.State != models.LobbyStateReadyingUp {
		return errors.New("Lobby hasn't been filled up yet.")
	}

	tperr = lobby.ReadyPlayer(player)

	if tperr != nil {
		return tperr
	}

	if lobby.IsEveryoneReady() {
		lobby.Start()

		hooks.BroadcastLobbyStart(lobby)
		models.BroadcastLobbyList()
	}

	return emptySuccess
}

func (Player) PlayerNotReady(so *wsevent.Client, _ struct{}) interface{} {
	player := chelpers.GetPlayer(so.Token)
	lobbyid, tperr := player.GetLobbyID(false)
	if tperr != nil {
		return tperr
	}

	lobby, tperr := models.GetLobbyByID(lobbyid)
	if tperr != nil {
		return tperr
	}

	if lobby.State != models.LobbyStateReadyingUp {
		return errors.New("Lobby hasn't been filled up yet.")
	}

	tperr = lobby.UnreadyPlayer(player)
	lobby.RemovePlayer(player)
	hooks.AfterLobbyLeave(lobby, player)
	if spec := sessions.IsSpectating(so.ID, lobby.ID); spec {
		// IsSpectating checks if the player has joined the lobby's public room
		lobby.AddSpectator(player)
	}

	if tperr != nil {
		return tperr
	}

	lobby.SetState(models.LobbyStateWaiting)
	lobby.UnreadyAllPlayers()
	models.BroadcastLobby(lobby)
	return emptySuccess
}

func (Player) PlayerSettingsGet(so *wsevent.Client, args struct {
	Key *string `json:"key"`
}) interface{} {

	player := chelpers.GetPlayer(so.Token)
	if *args.Key == "*" {
		return newResponse(player.Settings)
	}

	setting := player.GetSetting(*args.Key)
	return newResponse(setting)
}

var reMumbleNick = regexp.MustCompile(`\w+`)

func (Player) PlayerSettingsSet(so *wsevent.Client, args struct {
	Key   *string `json:"key"`
	Value *string `json:"value"`
}) interface{} {

	player := chelpers.GetPlayer(so.Token)

	switch *args.Key {
	case "siteAlias":
		if len(*args.Value) > 32 {
			return errors.New("Site alias must be under 32 characters long.")
		}

		player.SetSetting(*args.Key, *args.Value)

		player.SetPlayerProfile()
		so.EmitJSON(helpers.NewRequest("playerProfile", player))

		if lobbyID, _ := player.GetLobbyID(true); lobbyID != 0 {
			lobby, _ := models.GetLobbyByID(lobbyID)
			slot, _ := lobby.GetPlayerSlot(player)
			player.SetMumbleUsername(lobby.Type, slot)
			lobbyData := lobby.LobbyData(true)
			lobbyData.Send()
		}
	default:
		player.SetSetting(*args.Key, *args.Value)
	}

	return emptySuccess
}

func (Player) PlayerProfile(so *wsevent.Client, args struct {
	Steamid *string `json:"steamid"`
}) interface{} {

	steamid := *args.Steamid
	if steamid == "" {
		steamid = so.Token.Claims["steam_id"].(string)
	}

	player, err := models.GetPlayerBySteamID(steamid)
	if err != nil {
		return err
	}

	player.SetPlayerProfile()
	return newResponse(player)
}

var (
	changeMu = new(sync.RWMutex)
	//stores the last time the player made a change to
	//the twitch bot's status (leave/join their channel)
	lastTwitchBotChange = make(map[uint]time.Time)
)

func (Player) PlayerEnableTwitchBot(so *wsevent.Client, _ struct{}) interface{} {
	player := chelpers.GetPlayer(so.Token)
	if player.TwitchName == "" {
		return errors.New("Please connect your Twitch Account first.")
	}

	changeMu.RLock()
	last := lastTwitchBotChange[player.ID]
	changeMu.RUnlock()
	if time.Since(last) < time.Minute {
		return errors.New("Please wait for a minute before changing the bot's status")
	}

	models.TwitchBotJoin(player.TwitchName)

	changeMu.Lock()
	lastTwitchBotChange[player.ID] = time.Now()
	changeMu.Unlock()

	time.AfterFunc(1*time.Minute, func() {
		changeMu.Lock()
		delete(lastTwitchBotChange, player.ID)
		changeMu.Unlock()
	})

	return emptySuccess
}

func (Player) PlayerDisableTwitchBot(so *wsevent.Client, _ struct{}) interface{} {
	player := chelpers.GetPlayer(so.Token)
	if player.TwitchName == "" {
		return errors.New("Please connect your Twitch Account first.")
	}

	changeMu.RLock()
	last := lastTwitchBotChange[player.ID]
	changeMu.RUnlock()
	if time.Since(last) < time.Minute {
		return errors.New("Please wait for a minute before changing the bot's status")
	}

	models.TwitchBotLeave(player.TwitchName)

	changeMu.Lock()
	lastTwitchBotChange[player.ID] = time.Now()
	changeMu.Unlock()

	time.AfterFunc(1*time.Minute, func() {
		changeMu.Lock()
		delete(lastTwitchBotChange, player.ID)
		changeMu.Unlock()
	})
	return emptySuccess
}
