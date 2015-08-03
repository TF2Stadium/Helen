package socket

import (
	"html"
	"strconv"
	"strings"
	"time"

	chelpers "github.com/TF2Stadium/Server/controllers/controllerhelpers"
	"github.com/TF2Stadium/Server/database"
	"github.com/TF2Stadium/Server/helpers"
	"github.com/TF2Stadium/Server/models"
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

	so.On("authenticationTest", func(data string) string {
		var answer string
		if chelpers.IsLoggedInSocket(so.Id()) {
			answer = "authenticated"
		} else {
			answer = "not authenticated"
		}

		return answer
	})

	helpers.Logger.Debug("on connection")
	so.Join("-1") //room for global chat

	if chelpers.IsLoggedInSocket(so.Id()) {
		player, _ := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))
		lobbyid, err := player.GetLobbyId()
		if err != nil {
			so.Join(strconv.FormatUint(uint64(lobbyid), 10))
		}
	}

	so.On("lobbyCreate", func(jsonstr string) string {
		if !chelpers.IsLoggedInSocket(so.Id()) {
			bytes, _ := chelpers.BuildFailureJSON("Player isn't logged in.", -4).Encode()
			return string(bytes)
		}

		js, err := simplejson.NewFromReader(strings.NewReader(jsonstr))
		if err != nil {
			bytes, _ := chelpers.BuildFailureJSON("Malformed JSON syntax.", 0).Encode()
			return string(bytes)
		}

		mapName, err := js.Get("mapName").String()
		if err != nil {
			bytes, _ := chelpers.BuildMissingArgJSON("mapName").Encode()
			return string(bytes)
		}

		lobbytype, err := js.Get("type").String()
		if err != nil {
			bytes, _ := chelpers.BuildMissingArgJSON("type").Encode()
			return string(bytes)
		}

		server, err := js.Get("server").String()
		if err != nil {
			bytes, _ := chelpers.BuildMissingArgJSON("server").Encode()
			return string(bytes)
		}

		rconPwd, err := js.Get("rconpwd").String()
		if err != nil {
			bytes, _ := chelpers.BuildMissingArgJSON("rconpwn").Encode()
			return string(bytes)
		}

		whitelist, err := js.Get("whitelist").Int()
		if err != nil {
			bytes, _ := chelpers.BuildMissingArgJSON("whitelist").Encode()
			return string(bytes)
		}

		//mumble, _ := js.Get("mumbleRequired").Bool()

		var playermap = map[string]models.LobbyType{
			"sixes":      models.LobbyTypeSixes,
			"highlander": models.LobbyTypeHighlander,
		}

		//TODO: Configure server here

		//TODO what if playermap[lobbytype] is nil?
		lob := models.NewLobby(mapName, playermap[lobbytype],
			models.ServerRecord{Host: server, RconPassword: rconPwd}, whitelist)
		err = lob.Save()

		if err != nil {
			bytes, _ := err.(*helpers.TPError).ErrorJSON().Encode()
			return string(bytes)
		}

		lobby_id := simplejson.New()
		lobby_id.Set("id", lob.ID)
		bytes, _ := chelpers.BuildSuccessJSON(lobby_id).Encode()
		return string(bytes)
	})
	so.On("lobbyJoin", func(jsonstr string) string {
		if !chelpers.IsLoggedInSocket(so.Id()) {
			bytes, _ := chelpers.BuildFailureJSON("Player isn't logged in.", -4).Encode()
			return string(bytes)
		}
		js, err := simplejson.NewFromReader(strings.NewReader(jsonstr))

		if err != nil {
			bytes, _ := chelpers.BuildFailureJSON("Malformed JSON syntax.", 0).Encode()
			return string(bytes)
		}

		player, tperr := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))

		if tperr != nil {
			bytes, _ := tperr.ErrorJSON().Encode()
			return string(bytes)
		}

		lobbyid, err := js.Get("id").Uint64()
		if err != nil {
			bytes, _ := chelpers.BuildMissingArgJSON("id").Encode()
			return string(bytes)
		}
		lob, tperr := models.GetLobbyById(uint(lobbyid))
		if tperr != nil {
			bytes, _ := tperr.ErrorJSON().Encode()
			return string(bytes)
		}

		classString, err := js.Get("class").String()
		if err != nil {
			bytes, _ := chelpers.BuildMissingArgJSON("class").Encode()
			return string(bytes)
		}

		teamString, err := js.Get("team").String()
		if err != nil {
			bytes, _ := chelpers.BuildMissingArgJSON("team").Encode()
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
	})
	so.On("lobbyRemovePlayer", func(jsonstr string) string {
		if !chelpers.IsLoggedInSocket(so.Id()) {
			bytes, _ := chelpers.BuildFailureJSON("Player isn't logged in.", -4).Encode()
			return string(bytes)
		}

		js, err := simplejson.NewFromReader(strings.NewReader(jsonstr))
		if err != nil {
			bytes, _ := chelpers.BuildFailureJSON("Malformed JSON syntax.", 0).Encode()
			return string(bytes)
		}

		steamid, err := js.Get("steamid").String()
		// TODO check authorisation, currently can kick anyone
		if err != nil || steamid == "" {
			steamid = chelpers.GetSteamId(so.Id())
		}

		ban, err := js.Get("ban").Bool()
		if err != nil {
			bytes, _ := chelpers.BuildMissingArgJSON("ban").Encode()
			return string(bytes)
		}
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

		lob := &models.Lobby{}
		database.DB.Find(lob, lobbyid)

		if ban {
			lob.KickAndBanPlayer(player)
		} else {
			lob.RemovePlayer(player)
		}
		so.Leave(strconv.FormatInt(int64(lobbyid), 10))
		bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
		return string(bytes)
	})

	so.On("playerReady", func(jsonstr string) string {
		if !chelpers.IsLoggedInSocket(so.Id()) {
			bytes, _ := chelpers.BuildFailureJSON("Player isn't logged in.", -4).Encode()
			return string(bytes)
		}

		player, tperr := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))
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

		bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
		return string(bytes)
	})

	so.On("playerUnready", func(jsonstr string) string {
		if !chelpers.IsLoggedInSocket(so.Id()) {
			bytes, _ := chelpers.BuildFailureJSON("Player isn't logged in.", -4).Encode()
			return string(bytes)
		}

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

	so.On("lobbyJoinSpectator", func(jsonstr string) string {
		if !chelpers.IsLoggedInSocket(so.Id()) {
			bytes, _ := chelpers.BuildFailureJSON("Player isn't logged in.", -4).Encode()
			return string(bytes)
		}

		js, err := simplejson.NewFromReader(strings.NewReader(jsonstr))
		if err != nil {
			bytes, _ := chelpers.BuildFailureJSON("Malformed JSON syntax.", 0).Encode()
			return string(bytes)
		}
		lobbyid, err := js.Get("id").Uint64()
		if err != nil {
			bytes, _ := chelpers.BuildMissingArgJSON("id").Encode()
			return string(bytes)
		}
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
	})

	so.On("removeSpectator", func(jsonstr string) string {
		if !chelpers.IsLoggedInSocket(so.Id()) {
			bytes, _ := chelpers.BuildFailureJSON("Player isn't logged in.", -4).Encode()
			return string(bytes)
		}

		js, err := simplejson.NewFromReader(strings.NewReader(jsonstr))
		if err != nil {
			bytes, _ := chelpers.BuildFailureJSON("Malformed JSON syntax.", 0).Encode()
			return string(bytes)
		}
		lobbyid, err := js.Get("id").Uint64()
		if err != nil {
			bytes, _ := chelpers.BuildMissingArgJSON("id").Encode()
			return string(bytes)
		}
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
		tperr = lob.RemoveSpectator(player)
		if tperr != nil {
			bytes, _ := tperr.ErrorJSON().Encode()
			return string(bytes)
		}
		lob.Save()
		return string(bytes)
	})

	so.On("chatSend", func(jsonstr string) string {
		if !chelpers.IsLoggedInSocket(so.Id()) {
			bytes, _ := chelpers.BuildFailureJSON("Player isn't logged in.", -4).Encode()
			return string(bytes)
		}

		js, err := simplejson.NewFromReader(strings.NewReader(jsonstr))
		if err != nil {
			bytes, _ := chelpers.BuildFailureJSON("Malformed JSON syntax.", 0).Encode()
			return string(bytes)
		}
		message, err := js.Get("message").String()
		if err != nil {
			bytes, _ := chelpers.BuildMissingArgJSON("message").Encode()
			return string(bytes)
		}
		room, err := js.Get("room").Int()
		if err != nil {
			bytes, _ := chelpers.BuildMissingArgJSON("room").Encode()
			return string(bytes)
		}

		if err != nil {
			bytes, _ := helpers.NewTPError("room must be an integer", -1).ErrorJSON().Encode()
			return string(bytes)
		}

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
		so.BroadcastTo(strconv.Itoa(room), "chatReceive", string(bytes))

		resp, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
		return string(resp)
	})

}
