package socket

import (
	"log"
	"strconv"
	"strings"
	"time"

	chelpers "github.com/TF2Stadium/Server/controllers/controllerhelpers"
	"github.com/TF2Stadium/Server/database"
	"github.com/TF2Stadium/Server/models"
	"github.com/bitly/go-simplejson"
	"github.com/googollee/go-socket.io"
)

func SocketInit(so socketio.Socket) {
	chelpers.AuthenticateSocket(so.Id(), so.Request())

	so.On("disconnection", func() {
		// chelpers.DeauthenticateSocket(so.Id())
		log.Println("on disconnect")
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

	log.Println("on connection")
	so.Join("-1") //room for global chat
	so.On("lobbyCreate", func(jsonstr string) string {
		js, err := simplejson.NewFromReader(strings.NewReader(jsonstr))
		if err != nil {
			bytes, _ := chelpers.BuildFailureJSON("Malformed JSON syntax.", 0).Encode()
			return string(bytes)
		}

		mapName, _ := js.Get("mapName").String()
		format, _ := js.Get("format").String()
		//server, _ := js.Get("server").String()
		//rconPwd, _ := js.Get("rconpwd").String()
		whitelist, _ := js.Get("whitelist").Int()
		//mumble, _ := js.Get("mumbleRequired").Bool()

		var playermap = map[string]models.LobbyType{
			"sixes":      models.LobbyTypeSixes,
			"highlander": models.LobbyTypeHighlander,
		}

		//TODO: Configure server here
		lob := models.NewLobby(mapName, playermap[format], whitelist)
		err = lob.Save()

		if err != nil {
			//TODO: Add stuff here
		}

		lobby_id := simplejson.New()
		lobby_id.Set("id", string(lob.ID))
		bytes, _ := chelpers.BuildSuccessJSON(lobby_id).Encode()
		return string(bytes)
	})
	so.On("lobbyJoin", func(jsonstr string) string {
		js, err := simplejson.NewFromReader(strings.NewReader(jsonstr))
		if err != nil {
			bytes, _ := chelpers.BuildFailureJSON("Malformed JSON syntax.", 0).Encode()
			return string(bytes)
		}

		//TODO: Use websockets session code for getting Player
		//something like session.Values["steamid"]
		var player *models.Player

		slot, _ := js.Get("slot").Int()
		lobbyid, _ := js.Get("lobbyid").Uint64()
		var lob *models.Lobby
		var bytes []byte

		lob, tperr := models.GetLobbyById(uint(lobbyid))
		if tperr != nil {
			bytes, _ = tperr.ErrorJSON().Encode()
			return string(bytes)
		}

		tperr = lob.AddPlayer(player, slot)
		if tperr != nil {
			bytes, _ = tperr.ErrorJSON().Encode()
			return string(bytes)
		}
		so.Join(strconv.FormatUint(lobbyid, 10))
		bytes, _ = chelpers.BuildSuccessJSON(simplejson.New()).Encode()
		return string(bytes)
	})
	so.On("lobbyRemovePlayer", func(jsonstr string) string {
		js, err := simplejson.NewFromReader(strings.NewReader(jsonstr))
		if err != nil {
			bytes, _ := chelpers.BuildFailureJSON("Malformed JSON syntax.", 0).Encode()
			return string(bytes)
		}

		var steamid string
		var bytes []byte

		steamidjson, gotem := js.CheckGet("steamid")
		if !gotem {
			//Get SteamID of current player
			//TODO: Use websockets session code for getting Player
			//something like player := session.Values["steamid"]
		} else {
			steamid, _ = steamidjson.String()
		}
		ban, _ := js.Get("ban").Bool()
		player, tperr := models.GetPlayerBySteamId(steamid)

		if tperr != nil {
			bytes, _ = tperr.ErrorJSON().Encode()
			return string(bytes)
		}

		lobbyid, err := player.GetLobbyId()

		if err != nil {
			bytes, _ = chelpers.BuildFailureJSON("Player not in any Lobby.", 4).Encode()
			return string(bytes)
		}

		lob := &models.Lobby{}
		database.DB.Find(lob, lobbyid)

		if ban {
			lob.KickAndBanPlayer(player)
		} else {
			lob.RemovePlayer(player)
		}
		so.Leave(strconv.FormatUInt(uint(lobbyid), 10))
		bytes, _ = chelpers.BuildSuccessJSON(simplejson.New()).Encode()
		return string(bytes)
	})

	so.On("readyPlayer", func(jsonstr string) string {
		//Get SteamID of current player
		//TODO: Use websockets session code for getting Player
		//something like player := session.Values["steamid"]

		var steamid string
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

		bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
		return string(bytes)
	})

	so.On("unreadyPlayer", func(jsonstr string) string {
		//Get SteamID of current player
		//TODO: Use websockets session code for getting Player
		//something like player := session.Values["steamid"]
		var steamid string
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

		tperr = lobby.UnreadyPlayer(player)
		if tperr != nil {
			bytes, _ := tperr.ErrorJSON().Encode()
			return string(bytes)
		}

		bytes, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
		return string(bytes)
	})

	so.On("addSpectator", func(jsonstr string) string {
		js, err := simplejson.NewFromReader(strings.NewReader(jsonstr))
		if err != nil {
			bytes, _ := chelpers.BuildFailureJSON("Malformed JSON syntax.", 0).Encode()
			return string(bytes)
		}
		steamid, _ := js.Get("steamid").String()
		lobbyid, _ := js.Get("id").Uint64()
		player, tperr := models.GetPlayerBySteamId(steamid)
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
		js, err := simplejson.NewFromReader(strings.NewReader(jsonstr))
		if err != nil {
			bytes, _ := chelpers.BuildFailureJSON("Malformed JSON syntax.", 0).Encode()
			return string(bytes)
		}
		steamid, _ := js.Get("steamid").String()
		lobbyid, _ := js.Get("id").Uint64()
		player, tperr := models.GetPlayerBySteamId(steamid)
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
		js, err := simplejson.NewFromReader(strings.NewReader(jsonstr))
		if err != nil {
			bytes, _ := chelpers.BuildFailureJSON("Malformed JSON syntax.", 0).Encode()
			return string(bytes)
		}
		message, _ := js.Get("message").String()
		room, _ := js.Get("room").Int64()
		//TODO: Use websockets session code for getting Player
		//something like session.Values["steamid"]
		var player *models.Player
		//Check if player has either joined, or is spectating lobby
		lobbyId, tperr := player.GetLobbyId()
		if room != -1 {
			if tperr != nil {
				bytes, _ := tperr.ErrorJSON().Encode()
				return string(bytes)
			} else if lobbyId != uint(room) && !player.IsSpectatingId(uint(room)) {
				bytes, _ := chelpers.BuildFailureJSON("Player is not in the lobby.", 5).Encode()
				return string(bytes)
			}
		}
		t := time.Now()
		chatMessage := simplejson.New()
		chatMessage.Set("timestamp", strconv.Itoa(t.Hour())+strconv.Itoa(t.Minute()))
		chatMessage.Set("message", message)
		chatMessage.Set("room", room)

		user := simplejson.New()
		user.Set("id", player.SteamId)
		user.Set("name", player.Name)

		chatMessage.Set("user", user)
		bytes, _ := chatMessage.Encode()
		so.BroadcastTo(strconv.FormatInt(room, 10), string(bytes))

		resp, _ := chelpers.BuildSuccessJSON(simplejson.New()).Encode()
		return string(resp)
	})

}
