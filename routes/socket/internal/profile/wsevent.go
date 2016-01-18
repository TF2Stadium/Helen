package profile

import (
	"encoding/json"
	"net/http"

	"github.com/TF2Stadium/wsevent"
)

type roomProf struct {
	Room       string   `json:"room"`
	ClientsNum int      `json:"clients_num"`
	Clients    []string `json:"clients,omitempty"`
}

func serverProfile(server *wsevent.Server) []roomProf {
	rooms := server.Rooms()

	var serverProf []roomProf
	for room, clientCount := range rooms {
		serverProf = append(serverProf, roomProf{Room: room, ClientsNum: clientCount})
	}

	return serverProf
}

func Profile(server *wsevent.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bytes, _ := json.MarshalIndent(serverProfile(server), "", "  ")
		w.Write(bytes)
	}
}
