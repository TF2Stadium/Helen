package socket

import (
	"log"
	"strings"

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
	so.Join("chat")
	so.On("chat message", func(msg string) {
		log.Println("emit:", so.Emit("chat message", msg))
		so.BroadcastTo("chat", "chat message", msg)
	})

	so.On("lobbyCreate", func(jsonstr string) string {
		js, _ := simplejson.NewFromReader(strings.NewReader(jsonstr))

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
		err := lob.Save()

		if err != nil {
			//TODO: Add stuff here
		}

		lobby_id := simplejson.New()
		lobby_id.Set("id", string(lob.ID))
		bytes, _ := chelpers.BuildSuccessJSON(lobby_id).Encode()
		return string(bytes)
	})
	so.On("lobbyJoin", func(jsonstr string) string {
		js, _ := simplejson.NewFromReader(strings.NewReader(jsonstr))

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

		bytes, _ = chelpers.BuildSuccessJSON(simplejson.New()).Encode()
		return string(bytes)
	})
	so.On("lobbyRemovePlayer", func(jsonstr string) string {
		js, _ := simplejson.NewFromReader(strings.NewReader(jsonstr))

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
		lob.RemovePlayer(player)
		bytes, _ = chelpers.BuildSuccessJSON(simplejson.New()).Encode()
		return string(bytes)
	})

}
