package socket

import (
	"crypto/rand"
	"encoding/base64"
	"html"
	"strconv"
	"time"

	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/decorators"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/bitly/go-simplejson"
	"github.com/googollee/go-socket.io"
)

func SocketInit(so socketio.Socket) {
	chelpers.AuthenticateSocket(so.Id(), so.Request())
	if chelpers.IsLoggedInSocket(so.Id()) {
		steamid := chelpers.GetSteamId(so.Id())
		SteamIdSocketMap[steamid] = &so
	}

	so.On("disconnection", func() {
		chelpers.DeauthenticateSocket(so.Id())
		if chelpers.IsLoggedInSocket(so.Id()) {
			steamid := chelpers.GetSteamId(so.Id())
			delete(SteamIdSocketMap, steamid)
		}
		helpers.Logger.Debug("on disconnect")
	})

	so.On("authenticationTest", chelpers.AuthFilter(so.Id(), func(val string) string {
		return "authenticated"
	}))

	helpers.Logger.Debug("on connection")
	so.Join("-1") //room for global chat

	if chelpers.IsLoggedInSocket(so.Id()) {
		player, err := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))
		if err != nil {
			helpers.Logger.Warning("User has a cookie with but a matching player record doesn't exist: %s",
				chelpers.GetSteamId(so.Id()))
			return
		}
		lobbyid, err := player.GetLobbyId()
		if err != nil {
			so.Join(strconv.FormatUint(uint64(lobbyid), 10))
		}
	}

	var lobbyCreateParams = map[string]chelpers.Param{
		"mapName":        chelpers.Param{Type: chelpers.PTypeString},
		"type":           chelpers.Param{Type: chelpers.PTypeString},
		"league":         chelpers.Param{Type: chelpers.PTypeString},
		"server":         chelpers.Param{Type: chelpers.PTypeString},
		"rconpwd":        chelpers.Param{Type: chelpers.PTypeString},
		"whitelist":      chelpers.Param{Type: chelpers.PTypeInt},
		"mumbleRequired": chelpers.Param{Type: chelpers.PTypeBool},
	}

	so.On("lobbyCreate", chelpers.AuthFilter(so.Id(),
		chelpers.JsonVerifiedFilter(lobbyCreateParams, func(js *simplejson.Json) string {

			player, _ := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

			mapName, _ := js.Get("mapName").String()
			lobbytypestring, _ := js.Get("type").String()
			league, _ := js.Get("league").String()
			server, _ := js.Get("server").String()
			rconPwd, _ := js.Get("rconpwd").String()
			whitelist, err := js.Get("whitelist").Int()

			var playermap = map[string]models.LobbyType{
				"sixes":      models.LobbyTypeSixes,
				"highlander": models.LobbyTypeHighlander,
			}

			lobbytype, ok := playermap[lobbytypestring]
			if !ok {
				bytes, _ := chelpers.BuildFailureJSON("Lobby type invalid.", -1).Encode()
				return string(bytes)
			}
			if !models.IsLeagueValid(league) {
				bytes, _ := chelpers.BuildFailureJSON("Invalid League Name", -1).Encode()
				return string(bytes)
			}

			randBytes := make([]byte, 6)
			rand.Read(randBytes)
			serverPwd := base64.URLEncoding.EncodeToString(randBytes)
			//mumble, _ := js.Get("mumbleRequired").Bool()
			//TODO: Configure server here

			//TODO what if playermap[lobbytype] is nil?
			lob := models.NewLobby(mapName, lobbytype,
				models.ServerRecord{Host: server, RconPassword: rconPwd, ServerPassword: serverPwd}, whitelist)
			lob.CreatedBy = *player
			lob.Save()
			err = lob.SetupServer()
			if err != nil {
				bytes, _ := err.(*helpers.TPError).ErrorJSON().Encode()
				return string(bytes)
			}

			lobby_id := simplejson.New()
			lobby_id.Set("id", lob.ID)
			bytes, _ := chelpers.BuildSuccessJSON(lobby_id).Encode()
			return string(bytes)
		})))

	var lobbyCloseParams = map[string]chelpers.Param{
		"id": chelpers.Param{Type: chelpers.PTypeInt},
	}

	so.On("lobbyClose", chelpers.AuthFilter(so.Id(),
		chelpers.JsonVerifiedFilter(lobbyCloseParams, func(js *simplejson.Json) string {
			player, _ := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

			lobbyid, _ := js.Get("id").Uint64()

			lob, tperr := models.GetLobbyById(uint(lobbyid))
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			if player.ID != lob.CreatedByID {
				bytes, _ := chelpers.BuildFailureJSON("Player not authorized to close lobby.", 1).Encode()
				return string(bytes)
			}

			if lob.State == models.LobbyStateEnded {
				bytes, _ := chelpers.BuildFailureJSON("Lobby already closed.", -1).Encode()
				return string(bytes)
			}

			lob.Close(true)

			bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
			return string(bytes)
		})))

	var lobbyJoinParams = map[string]chelpers.Param{
		"id":    chelpers.Param{Type: chelpers.PTypeInt},
		"class": chelpers.Param{Type: chelpers.PTypeString},
		"team":  chelpers.Param{Type: chelpers.PTypeString},
	}

	so.On("lobbyJoin", chelpers.AuthFilter(so.Id(),
		chelpers.JsonVerifiedFilter(lobbyJoinParams, func(js *simplejson.Json) string {
			player, tperr := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			lobbyid, _ := js.Get("id").Uint64()
			classString, _ := js.Get("class").String()
			teamString, _ := js.Get("team").String()

			lob, tperr := models.GetLobbyById(uint(lobbyid))
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			slot, tperr := chelpers.GetPlayerSlot(lob.Type, teamString, classString)
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			tperr = lob.AddPlayer(player, slot)
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			so.Join(strconv.FormatUint(lobbyid, 10))
			bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
			return string(bytes)
		})))

	var lobbyRemovePlayerParams = map[string]chelpers.Param{
		"id":      chelpers.Param{Type: chelpers.PTypeInt},
		"steamid": chelpers.Param{Type: chelpers.PTypeString, Default: ""},
		"ban":     chelpers.Param{Type: chelpers.PTypeBool, Default: false},
	}

	so.On("lobbyRemovePlayer", chelpers.AuthFilter(so.Id(),
		chelpers.JsonVerifiedFilter(lobbyRemovePlayerParams, func(js *simplejson.Json) string {
			steamid, _ := js.Get("steamid").String()
			ban, _ := js.Get("ban").Bool()
			lobbyid, _ := js.Get("id").Int()
			self := false

			// TODO check authorization, currently can kick anyone
			if steamid == "" || steamid == chelpers.GetSteamId(so.Id()) {
				self = true
				steamid = chelpers.GetSteamId(so.Id())
			}

			player, tperr := models.GetPlayerBySteamId(steamid)
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			lob, tperr := models.GetLobbyById(uint(lobbyid))
			if tperr != nil {
				bytes, _ := chelpers.BuildFailureJSON(tperr.Error(), -1).Encode()
				return string(bytes)
			}

			if !self && lob.CreatedByID != player.ID {
				// TODO proper authorization checks
				bytes, _ := chelpers.BuildFailureJSON("Not authorized to remove players", 1).Encode()
				return string(bytes)
			}

			_, err := lob.GetPlayerSlot(player)
			if err == nil {
				lob.RemovePlayer(player)
			} else if player.IsSpectatingId(lob.ID) {
				lob.RemoveSpectator(player)
			} else {
				bytes, _ := chelpers.BuildFailureJSON("Player neither playing nor spectating", 2).Encode()
				return string(bytes)
			}

			if ban {
				lob.BanPlayer(player)
			}

			so.Leave(strconv.FormatInt(int64(lobbyid), 10))
			bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
			return string(bytes)
		})))

	so.On("playerReady", chelpers.AuthFilter(so.Id(), func(val string) string {
		steamid := chelpers.GetSteamId(so.Id())
		player, tperr := models.GetPlayerBySteamId(steamid)
		if tperr != nil {
			bytes, _ := tperr.ErrorJSON().Encode()
			return string(bytes)
		}

		lobbyid, tperr := player.GetLobbyId()
		if tperr != nil {
			bytes, _ := tperr.ErrorJSON().Encode()
			return string(bytes)
		}

		lobby, tperr := models.GetLobbyById(lobbyid)
		if tperr != nil {
			bytes, _ := tperr.ErrorJSON().Encode()
			return string(bytes)
		}

		tperr = lobby.ReadyPlayer(player)
		if tperr != nil {
			bytes, _ := tperr.ErrorJSON().Encode()
			return string(bytes)
		}

		if lobby.IsEveryoneReady() {
			bytes, _ := decorators.GetLobbyConnectJSON(lobby).Encode()
			SendMessageToRoom(strconv.FormatUint(uint64(lobby.ID), 10),
				"lobbyStart", string(bytes))
		}

		bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
		return string(bytes)
	}))

	so.On("playerUnready", chelpers.AuthFilter(so.Id(), func(val string) string {
		player, tperr := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))
		if tperr != nil {
			bytes, _ := tperr.ErrorJSON().Encode()
			return string(bytes)
		}

		if tperr != nil {
			bytes, _ := tperr.ErrorJSON().Encode()
			return string(bytes)
		}

		lobbyid, tperr := player.GetLobbyId()
		if tperr != nil {
			bytes, _ := tperr.ErrorJSON().Encode()
			return string(bytes)
		}

		lobby, tperr := models.GetLobbyById(lobbyid)
		if tperr != nil {
			bytes, _ := tperr.ErrorJSON().Encode()
			return string(bytes)
		}

		tperr = lobby.UnreadyPlayer(player)
		if tperr != nil {
			bytes, _ := tperr.ErrorJSON().Encode()
			return string(bytes)
		}

		bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
		return string(bytes)
	}))

	var lobbyJoinSpectatorParams = map[string]chelpers.Param{
		"id": chelpers.Param{Type: chelpers.PTypeInt},
	}

	so.On("lobbySpectatorJoin", chelpers.AuthFilter(so.Id(),
		chelpers.JsonVerifiedFilter(lobbyJoinSpectatorParams, func(js *simplejson.Json) string {
			lobbyid, _ := js.Get("id").Uint64()

			player, tperr := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}
			lob, tperr := models.GetLobbyById(uint(lobbyid))
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}
			bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
			tperr = lob.AddSpectator(player)
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}
			lob.Save()
			return string(bytes)
		})))

	var playerSettingsGetParams = map[string]chelpers.Param{
		"key": chelpers.Param{Type: chelpers.PTypeString, Default: ""},
	}

	so.On("playerSettingsGet", chelpers.AuthFilter(so.Id(),
		chelpers.JsonVerifiedFilter(playerSettingsGetParams, func(js *simplejson.Json) string {
			player, _ := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

			key, _ := js.Get("key").String()

			var err error
			var settings []models.PlayerSetting
			var setting models.PlayerSetting
			if key == "" {
				settings, err = player.GetSettings()
			} else {
				setting, err = player.GetSetting(key)
				settings = append(settings, setting)
			}

			if err != nil {
				bytes, _ := chelpers.BuildFailureJSON(err.Error(), 0).Encode()
				return string(bytes)
			}

			result := decorators.GetPlayerSettingsJson(settings)
			resp, _ := chelpers.BuildSuccessJSON(result).Encode()
			return string(resp)
		})))

	var playerSettingsSetParams = map[string]chelpers.Param{
		"key":   chelpers.Param{Type: chelpers.PTypeString},
		"value": chelpers.Param{Type: chelpers.PTypeString},
	}

	so.On("playerSettingsSet", chelpers.AuthFilter(so.Id(),
		chelpers.JsonVerifiedFilter(playerSettingsSetParams, func(js *simplejson.Json) string {
			player, _ := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

			key, _ := js.Get("key").String()
			value, _ := js.Get("value").String()

			err := player.SetSetting(key, value)
			if err != nil {
				bytes, _ := chelpers.BuildFailureJSON(err.Error(), 0).Encode()
				return string(bytes)
			}

			resp, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
			return string(resp)
		})))

	var playerProfileParams = map[string]chelpers.Param{
		"steamid": chelpers.Param{Type: chelpers.PTypeString, Default: ""},
	}

	so.On("playerProfile", chelpers.AuthFilter(so.Id(),
		chelpers.JsonVerifiedFilter(playerProfileParams, func(js *simplejson.Json) string {
			steamid, _ := js.Get("steamid").String()

			if steamid == "" {
				steamid = chelpers.GetSteamId(so.Id())
			}

			player, playErr := models.GetPlayerWithStats(steamid)

			if playErr != nil {
				bytes, _ := chelpers.BuildFailureJSON(playErr.Error(), 0).Encode()
				return string(bytes)
			}

			result := decorators.GetPlayerProfileJson(player)
			resp, _ := chelpers.BuildSuccessJSON(result).Encode()
			return string(resp)
		})))

	var chatSendParams = map[string]chelpers.Param{
		"message": chelpers.Param{Type: chelpers.PTypeString},
		"room":    chelpers.Param{Type: chelpers.PTypeInt},
	}

	so.On("chatSend", chelpers.AuthFilter(so.Id(),
		chelpers.JsonVerifiedFilter(chatSendParams, func(js *simplejson.Json) string {
			message, _ := js.Get("message").String()
			room, _ := js.Get("room").Int()

			player, tperr := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			//Check if player has either joined, or is spectating lobby
			lobbyId, tperr := player.GetLobbyId()
			if room > 0 {
				// if room is a lobby room
				if tperr != nil {
					bytes, _ := tperr.ErrorJSON().Encode()
					return string(bytes)
				} else if lobbyId != uint(room) && !player.IsSpectatingId(uint(room)) {
					bytes, _ := chelpers.BuildFailureJSON("Player is not in the lobby.", 5).Encode()
					return string(bytes)
				}
			} else {
				// else room is the lobby list room
				room = -1
			}

			t := time.Now()
			chatMessage := simplejson.New()
			// TODO send proper timestamps
			chatMessage.Set("timestamp", strconv.Itoa(t.Hour())+strconv.Itoa(t.Minute()))
			chatMessage.Set("message", html.EscapeString(message))
			chatMessage.Set("room", room)

			user := simplejson.New()
			user.Set("id", player.SteamId)
			user.Set("name", player.Name)

			chatMessage.Set("user", user)
			bytes, _ := chatMessage.Encode()
			socketServer.BroadcastTo(strconv.Itoa(room), "chatReceive", string(bytes))

			resp, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
			return string(resp)
		})))

	so.On("AdminChangeRole", ChangeRole(so.Id()))
}
