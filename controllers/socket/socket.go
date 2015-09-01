package socket

import (
	"crypto/rand"
	"encoding/base64"
	"html"
	"reflect"
	"strconv"
	"time"

	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/helpers/authority"
	"github.com/TF2Stadium/Helen/models"
	"github.com/bitly/go-simplejson"
	"github.com/googollee/go-socket.io"
)

func SocketInit(so socketio.Socket) {
	chelpers.AuthenticateSocket(so.Id(), so.Request())
	if chelpers.IsLoggedInSocket(so.Id()) {
		steamid := chelpers.GetSteamId(so.Id())
		broadcaster.SteamIdSocketMap[steamid] = &so
	}

	so.On("disconnection", func() {
		chelpers.DeauthenticateSocket(so.Id())
		if chelpers.IsLoggedInSocket(so.Id()) {
			steamid := chelpers.GetSteamId(so.Id())
			delete(broadcaster.SteamIdSocketMap, steamid)
		}
		helpers.Logger.Debug("on disconnect")
	})

	chelpers.RegisterEvent(so, "authenticationTest", chelpers.FilterParams{},
		func(_ map[string]interface{}) string {
			return "authenticated"
		})

	helpers.Logger.Debug("on connection")
	chelpers.AfterConnect(so)

	if chelpers.IsLoggedInSocket(so.Id()) {
		player, err := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))
		if err != nil {
			helpers.Logger.Warning("User has a cookie with but a matching player record doesn't exist: %s",
				chelpers.GetSteamId(so.Id()))
			return
		}

		chelpers.AfterConnectLoggedIn(so, player)
	}

	lobbyCreateFilters := chelpers.FilterParams{
		Action:      authority.AuthAction(0),
		FilterLogin: true,

		Params: map[string]chelpers.Param{
			"mapName": chelpers.Param{Kind: reflect.String},

			"type": chelpers.Param{
				Kind: reflect.String,
				In:   []string{"highlander", "sixes"}},
			"league": chelpers.Param{
				Kind: reflect.String,
				In:   []string{"etf2l", "ugc"}},

			"server": chelpers.Param{Kind: reflect.String},

			"rconpwd":        chelpers.Param{Kind: reflect.String},
			"whitelist":      chelpers.Param{Kind: reflect.Uint},
			"mumbleRequired": chelpers.Param{Kind: reflect.Bool},
		},
	}

	chelpers.RegisterEvent(so, "lobbyCreate", lobbyCreateFilters,
		func(params map[string]interface{}) string {

			player, _ := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

			mapName := params["mapName"].(string)
			lobbytypestring := params["type"].(string)
			league := params["league"].(string)
			server := params["server"].(string)
			rconPwd := params["rconpwd"].(string)
			whitelist := int(params["whitelist"].(uint))
			//mumble := params["mumbleRequired"].(bool)

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

			//TODO what if playermap[lobbytype] is nil?
			info := models.ServerRecord{
				Host:           server,
				RconPassword:   rconPwd,
				ServerPassword: serverPwd}
			err := models.VerifyInfo(info)
			if err != nil {
				return err.Error()
			}

			lob := models.NewLobby(mapName, lobbytype, league, info, whitelist)
			lob.CreatedBy = *player
			lob.Save()
			err = lob.SetupServer()

			if err != nil {
				bytes, _ := err.(*helpers.TPError).ErrorJSON().Encode()
				return string(bytes)
			}

			lob.State = models.LobbyStateWaiting
			lob.Save()
			lobby_id := simplejson.New()
			lobby_id.Set("id", lob.ID)
			bytes, _ := chelpers.BuildSuccessJSON(lobby_id).Encode()
			return string(bytes)
		})

	lobbyCloseFilters := chelpers.FilterParams{
		Action:      authority.AuthAction(0),
		FilterLogin: true,
		Params: map[string]chelpers.Param{
			"id": chelpers.Param{Kind: reflect.Uint},
		},
	}

	chelpers.RegisterEvent(so, "lobbyClose", lobbyCloseFilters,
		func(params map[string]interface{}) string {
			player, _ := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

			lobbyid := params["id"].(uint)

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
			models.BroadcastLobbyList() // has to be done manually for now

			bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
			return string(bytes)
		})

	lobbyJoinFilters := chelpers.FilterParams{
		Action:      authority.AuthAction(0),
		FilterLogin: true,
		Params: map[string]chelpers.Param{
			"id":    chelpers.Param{Kind: reflect.Uint},
			"class": chelpers.Param{Kind: reflect.String},
			"team":  chelpers.Param{Kind: reflect.String},
		},
	}

	chelpers.RegisterEvent(so, "lobbyJoin", lobbyJoinFilters,
		func(params map[string]interface{}) string {
			player, tperr := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			lobbyid := params["id"].(uint)
			classString := params["class"].(string)
			teamString := params["team"].(string)

			lob, tperr := models.GetLobbyById(uint(lobbyid))
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			slot, tperr := models.LobbyGetPlayerSlot(lob.Type, teamString, classString)
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			tperr = lob.AddPlayer(player, slot)
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			chelpers.AfterLobbyJoin(so, lob, player)

			bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
			return string(bytes)
		})

	lobbyRemovePlayerFilters := chelpers.FilterParams{
		Action:      authority.AuthAction(0),
		FilterLogin: true,
		Params: map[string]chelpers.Param{
			"id":      chelpers.Param{Kind: reflect.Uint},
			"steamid": chelpers.Param{Kind: reflect.String, Default: ""},
			"ban":     chelpers.Param{Kind: reflect.Bool, Default: false},
		},
	}

	chelpers.RegisterEvent(so, "lobbyRemovePlayer", lobbyRemovePlayerFilters,
		func(params map[string]interface{}) string {
			steamid := params["steamid"].(string)
			ban := params["ban"].(bool)
			lobbyid := params["id"].(uint)
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

			chelpers.AfterLobbyLeave(so, lob, player)
			bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
			return string(bytes)
		})

	playerReadyFilter := chelpers.FilterParams{
		Action:      authority.AuthAction(0),
		FilterLogin: true,
	}
	chelpers.RegisterEvent(so, "playerReady", playerReadyFilter,
		func(_ map[string]interface{}) string {
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
				bytes, _ := models.DecorateLobbyConnectJSON(lobby).Encode()
				broadcaster.SendMessageToRoom(strconv.FormatUint(uint64(lobby.ID), 10),
					"lobbyStart", string(bytes))
			}

			bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
			return string(bytes)
		})

	playerUnreadyFilter := chelpers.FilterParams{
		Action:      authority.AuthAction(0),
		FilterLogin: true,
	}
	chelpers.RegisterEvent(so, "playerUnready", playerUnreadyFilter,
		func(_ map[string]interface{}) string {
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
		})

	lobbySpectatorJoinFilter := chelpers.FilterParams{
		FilterLogin: true,
		Params: map[string]chelpers.Param{
			"id": chelpers.Param{Kind: reflect.Uint},
		},
	}

	chelpers.RegisterEvent(so, "lobbySpectatorJoin", lobbySpectatorJoinFilter,
		func(params map[string]interface{}) string {

			lobbyid := params["id"].(uint)

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

			chelpers.AfterLobbyJoin(so, lob, player)
			return string(bytes)
		})

	playerSettingsGetFilter := chelpers.FilterParams{
		FilterLogin: true,
		Params: map[string]chelpers.Param{
			"key": chelpers.Param{Kind: reflect.String, Default: ""},
		},
	}
	chelpers.RegisterEvent(so, "playerSettingsGet", playerSettingsGetFilter,
		func(params map[string]interface{}) string {
			player, _ := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

			key := params["key"].(string)

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

			result := models.DecoratePlayerSettingsJson(settings)
			resp, _ := chelpers.BuildSuccessJSON(result).Encode()
			return string(resp)
		})

	playerSettingsSetFilter := chelpers.FilterParams{
		FilterLogin: true,
		Params: map[string]chelpers.Param{
			"key":   chelpers.Param{Kind: reflect.String},
			"value": chelpers.Param{Kind: reflect.String},
		},
	}

	chelpers.RegisterEvent(so, "playerSettingsSet", playerSettingsSetFilter,
		func(params map[string]interface{}) string {
			player, _ := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

			key := params["key"].(string)
			value := params["value"].(string)

			err := player.SetSetting(key, value)
			if err != nil {
				bytes, _ := chelpers.BuildFailureJSON(err.Error(), 0).Encode()
				return string(bytes)
			}

			resp, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
			return string(resp)
		})

	playerProfileFilter := chelpers.FilterParams{
		FilterLogin: true,
		Params: map[string]chelpers.Param{
			"steamid": chelpers.Param{Kind: reflect.String, Default: ""},
		},
	}

	chelpers.RegisterEvent(so, "playerProfile", playerProfileFilter,
		func(params map[string]interface{}) string {

			steamid := params["steamid"].(string)

			if steamid == "" {
				steamid = chelpers.GetSteamId(so.Id())
			}

			player, playErr := models.GetPlayerWithStats(steamid)

			if playErr != nil {
				bytes, _ := chelpers.BuildFailureJSON(playErr.Error(), 0).Encode()
				return string(bytes)
			}

			result := models.DecoratePlayerProfileJson(player)
			resp, _ := chelpers.BuildSuccessJSON(result).Encode()
			return string(resp)
		})

	chatSendFilter := chelpers.FilterParams{
		FilterLogin: true,
		Params: map[string]chelpers.Param{
			"message": chelpers.Param{Kind: reflect.String},
			"room":    chelpers.Param{Kind: reflect.Int},
		},
	}

	chelpers.RegisterEvent(so, "chatSend", chatSendFilter,
		func(params map[string]interface{}) string {
			message := params["message"].(string)
			room := params["room"].(int)

			player, tperr := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			}

			//Check if player has either joined, or is spectating lobby
			lobbyId, tperr := player.GetLobbyId()
			if room >= 0 {
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
			broadcaster.SendMessageToRoom(strconv.Itoa(room), "chatReceive", string(bytes))

			resp, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
			return string(resp)
		})

	adminChangeRoleFilter := chelpers.FilterParams{
		Action:      helpers.ActionChangeRole,
		FilterLogin: true,
		Params: map[string]chelpers.Param{
			"steamid": chelpers.Param{Kind: reflect.String},
			"role":    chelpers.Param{Kind: reflect.String},
		},
	}
	chelpers.RegisterEvent(so, "adminChangeRole", adminChangeRoleFilter,
		func(params map[string]interface{}) string {
			return ChangeRole(&so, params["role"].(string), params["steamid"].(string))
		})

	so.On("requestLobbyListData", func(s string) string {
		models.BroadcastLobbyList()

		resp, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
		return string(resp)
	})

}
