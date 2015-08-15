package models

import (
	"github.com/TF2Stadium/Helen/helpers"
	"sync"

	"github.com/TF2Stadium/Helen/config"
	"github.com/bitly/go-simplejson"
	zmq "github.com/pebbe/zmq4"
)

var sock *zmq.Socket
var sockMutex = &sync.Mutex{}

func PaulingConnect() {
	ctx, err := zmq.NewContext()
	if err != nil {
		helpers.Logger.Fatal(err)
	}
	sock, err = ctx.NewSocket(zmq.REQ)
	if err != nil {
		helpers.Logger.Fatal(err)
	}

	helpers.Logger.Debug("Connecting to Pauling on %s", config.Constants.PaulingEndpoint)
	err = sock.Connect(config.Constants.PaulingEndpoint)
	if err != nil {
		helpers.Logger.Fatal(err)
	}
	helpers.Logger.Debug("Connected.")

	//Test connection
	sock.Send("{\"request\":\"test\"}", 0)
	_, err = sock.Recv(0)
	if err != nil {
		helpers.Logger.Fatal(err)
	}
}

func SendJSON(json *simplejson.Json) *simplejson.Json {
	bytes, _ := json.Encode()

	sockMutex.Lock()
	_, err := sock.SendBytes(bytes, 0)
	if err != nil {
		helpers.Logger.Fatal(err)
	}

	respBytes, err := sock.RecvBytes(0)
	if err != nil {
		helpers.Logger.Fatal(err)
	}
	sockMutex.Unlock()

	respJson := simplejson.New()
	respJson.UnmarshalJSON(respBytes)
	return respJson
}

func ReqSetupServer(id uint, info ServerRecord, matchType LobbyType, mapName string) *simplejson.Json {
	json := simplejson.New()
	json.Set("request", "setupServer")
	json.Set("id", id)
	json.Set("server", info.Host)
	json.Set("rconPwn", info.RconPassword)
	json.Set("matchType", matchType)
	json.Set("mapName", mapName)
	json.Set("serverPwd", info.ServerPassword)

	return SendJSON(json)
}

func ReqAddAllowedPlayer(id uint, steamid string) *simplejson.Json {
	json := simplejson.New()
	json.Set("request", "addAllowedPlayer")
	json.Set("id", id)
	json.Set("steamid", steamid)

	return SendJSON(json)
}

func ReqDisallowPlayer(id uint, steamid string) *simplejson.Json {
	json := simplejson.New()
	json.Set("request", "removeAllowedPlayer")
	json.Set("id", id)
	json.Set("steamid", steamid)

	return SendJSON(json)
}

func ReqGetEvent() *simplejson.Json {
	json := simplejson.New()
	json.Set("request", "getEvent")

	return SendJSON(json)
}

func ReqEnd(id uint) *simplejson.Json {
	json := simplejson.New()
	json.Set("request", "end")
	json.Set("id", id)

	return SendJSON(json)
}
