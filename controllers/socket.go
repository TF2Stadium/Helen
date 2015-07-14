package controllers

import (
	"log"
	"strings"

	"github.com/TeamPlayTF/Server/database"
	"github.com/TeamPlayTF/Server/models"
	"github.com/TeamPlayTF/Server/models/lobby"
	"github.com/bitly/go-simplejson"
	"github.com/googollee/go-socket.io"
	"gopkg.in/mgo.v2/bson"
)

func SocketInit(so socketio.Socket) {
	log.Println("on connection")
	so.Join("chat")
	so.On("chat message", func(msg string) {
		log.Println("emit:", so.Emit("chat message", msg))
		so.BroadcastTo("chat", "chat message", msg)
	})
	so.On("disconnection", func() {
		log.Println("on disconnect")
	})
	so.On("createLobby", func(jsonstr string, response func(interface{}) interface{}) {
		js, _ := simplejson.NewFromReader(strings.NewReader(jsonstr))

		mapName, _ := js.Get("mapName").String()
		format, _ := js.Get("format").String()
		//server, _ := js.Get("server").String()
		//rconPwd, _ := js.Get("rconpwd").String()
		whitelist, _ := js.Get("whitelist").Int()

		var playermap = map[string]lobby.LobbyType{
			"sixes":      lobby.LobbyTypeSixes,
			"highlander": lobby.LobbyTypeHighlander,
		}

		//TODO: Configure server here
		lob := lobby.New(mapName, playermap[format], whitelist)
		err := lob.Save()

		if err != nil {
			//TODO: Add stuff here
		}

		lobby_id := simplejson.New()
		lobby_id.Set("id", string(lob.Id))
		bytes, _ := buildSuccessJSON(lobby_id).Encode()
		response(string(bytes))
	})
	so.On("addPlayer", func(jsonstr string, response func(interface{}) interface{}) {
		js, _ := simplejson.NewFromReader(strings.NewReader(jsonstr))

		//TODO: Use websockets session code for getting Player
		//something like session.Values["steamid"]
		var player *models.Player

		slot, _ := js.Get("slot").Int()
		lobbyidstring, _ := js.Get("lobbyid").String()
		var lob *lobby.Lobby
		var bytes []byte

		//schalla is evil, assume he'll send us all sorts of stuff
		//check lobbyidstring
		if !bson.IsObjectIdHex(lobbyidstring) {
			bytes, _ = buildFailureJSON("lobbyid is not a valid hex representation", -2).Encode()
		} else {
			lobbyid := bson.ObjectId(lobbyidstring)
			err := database.GetLobbiesCollection().FindId(lobbyid).One(&lob)

			if err != nil {
				bytes, _ = buildFailureJSON("Lobby not in the database", -1).Encode()
			} else {
				err := lob.AddPlayer(player, slot)
				if err != nil {
					bytes, _ = err.ErrorJSON().Encode()
				} else {
					bytes, _ = buildSuccessJSON(simplejson.New()).Encode()
					lob.Save()
				}
			}
		}
		response(string(bytes))
	})
	so.On("removePlayer", func(jsonstr string, response func(interface{}) interface{}) {
		js, _ := simplejson.NewFromReader(strings.NewReader(jsonstr))

		var player *models.Player
		var steamid string
		var bytes []byte

		steamidjson, gotem := js.CheckGet("steamid")
		if !gotem {
			//Remove the current player
			//TODO: Use websockets session code for getting Player
			//something like player := session.Values["steamid"]
		} else {
			steamid, _ = steamidjson.String()
			err := database.GetPlayersCollection().Find(bson.M{"steamid": steamid}).One(&player)
			if err != nil {
				bytes, _ = buildFailureJSON("Player not in the database", -1).Encode()
			} else {
				lobbyid, err := player.InLobby()
				if err != nil {
					bytes, _ = buildFailureJSON("Player not in any Lobby.", 4).Encode()
				} else {
					var lob *lobby.Lobby
					database.GetLobbiesCollection().FindId(lobbyid).One(&lob)
					lob.RemovePlayer(player)
					lob.Save()
					bytes, _ = buildSuccessJSON(simplejson.New()).Encode()
				}
			}
			response(string(bytes))
		}
	})

}
