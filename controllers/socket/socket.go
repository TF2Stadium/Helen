package socket

import (
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

	so.On("disconnection", func() {
		// chelpers.DeauthenticateSocket(so.Id())
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

		mapName, _ := js.Get("mapName").String()
		lobbytype, _ := js.Get("type").String()
		server, _ := js.Get("server").String()
		rconPwd, _ := js.Get("rconpwd").String()
		whitelist, _ := js.Get("whitelist").Int()
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

		lobbyid, _ := js.Get("id").Uint64()
		lob, tperr := models.GetLobbyById(uint(lobbyid))
		if tperr != nil {
			bytes, _ := tperr.ErrorJSON().Encode()
			return string(bytes)
		}

		classString, _ := js.Get("class").String()
		teamString, _ := js.Get("team").String()

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

		ban, _ := js.Get("ban").Bool()
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
		message, _ := js.Get("message").String()
		room, err := js.Get("room").Int()

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
		chatMessage.Set("message", message)
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
