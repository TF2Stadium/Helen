package handler

import (
	"errors"
	"regexp"
	"sync"
	"time"

	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/controllers/controllerhelpers/hooks"
	"github.com/TF2Stadium/Helen/controllers/socket/sessions"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models/lobby"
	"github.com/TF2Stadium/Helen/models/player"
	"github.com/TF2Stadium/Helen/models/rpc"
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

	lob, err := lobby.GetLobbyByIDServer(lobbyid)
	if err != nil {
		return err
	}

	if lob.State != lobby.ReadyingUp {
		return errors.New("Lobby hasn't been filled up yet.")
	}

	err = lob.ReadyPlayer(player)

	if err != nil {
		return err
	}

	if lob.IsEveryoneReady() {
		lob.Start()

		hooks.BroadcastLobbyStart(lob)
		lobby.BroadcastLobbyList()
	}

	return emptySuccess
}

func (Player) PlayerNotReady(so *wsevent.Client, _ struct{}) interface{} {
	player := chelpers.GetPlayer(so.Token)
	lobbyid, tperr := player.GetLobbyID(false)
	if tperr != nil {
		return tperr
	}

	lob, err := lobby.GetLobbyByID(lobbyid)
	if err != nil {
		return err
	}

	if lob.State != lobby.ReadyingUp {
		return errors.New("Lobby hasn't been filled up yet.")
	}

	err = lob.UnreadyPlayer(player)
	lob.RemovePlayer(player)
	hooks.AfterLobbyLeave(lob, player, false, false)
	if spec := sessions.IsSpectating(so.ID, lob.ID); spec {
		// IsSpectating checks if the player has joined the lobby's public room
		lob.AddSpectator(player)
	}

	if tperr != nil {
		return tperr
	}

	lob.SetState(lobby.Waiting)
	lob.UnreadyAllPlayers()
	lobby.BroadcastLobby(lob)
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
			lob, _ := lobby.GetLobbyByID(lobbyID)
			slot, _ := lob.GetPlayerSlot(player)
			player.SetMumbleUsername(lob.Type, slot)
			lobby.BroadcastLobby(lob)
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

	player, err := player.GetPlayerBySteamID(steamid)
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

	rpc.TwitchBotJoin(player.TwitchName)

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

	rpc.TwitchBotLeave(player.TwitchName)

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

func (Player) PlayerRecentLobbies(so *wsevent.Client, args struct {
	SteamID *string `json:"steamid"`
	Lobbies *int    `json:"lobbies"`
	LobbyID int     `json:"lobbyId"` // start from this lobbyID, 0 when not specified in json
}) interface{} {
	var p *player.Player

	if *args.SteamID != "" {
		var err error
		p, err = player.GetPlayerBySteamID(*args.SteamID)
		if err != nil {
			return err
		}

	} else {
		p = chelpers.GetPlayer(so.Token)
	}

	var lobbies []*lobby.Lobby

	db.DB.Model(&lobby.Lobby{}).Joins("INNER JOIN lobby_slots ON lobbies.ID = lobby_slots.lobby_id").
		Where("lobbies.match_ended = TRUE and lobby_slots.player_id = ? AND lobby_slots.needs_sub = FALSE AND lobbies.ID >= ?", p.ID, args.LobbyID).
		Order("lobbies.id desc").
		Limit(*args.Lobbies).
		Find(&lobbies)

	return newResponse(lobby.DecorateLobbyListData(lobbies, true))
}
