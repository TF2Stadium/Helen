package lobby

import (
	"log"
	"time"

	"github.com/TeamPlayTF/PlayerStatsScraper/steamid"
	"github.com/TeamPlayTF/TF2RconWrapper"
	"gopkg.in/mgo.v2/bson"
)

type Server struct {
	Map  string // lobby map
	Name string // server name
	Rcon *TF2RconWrapper.TF2RconConnection

	League League
	Type   LobbyType // 9v9 6v6 4v4...

	Address string // server ip:port
	LobbyId bson.ObjectId

	Players        []TF2RconWrapper.Player // current number of players in the server
	AllowedPlayers []TF2RconWrapper.Player

	Config *ServerConfig // config that should run before the lobby starts
	Ticker verifyTicker  // timer that will verify()

	//ChatListener  *TF2RconWrapper.RconChatListener
	RconPassword  string // will store the rcon password specified by the client
	LobbyPassword string // will store the new server password from the lobby
}

// timer used in verify()
type verifyTicker struct {
	Ticker *time.Ticker
	Quit   chan struct{}
}

func (t *verifyTicker) Close() {
	close(t.Quit)
}

func NewServer() *Server {
	return new(Server)
}

// after create the server var, you should run this
//
// things that needs to be specified before run this:
// -> Map
// -> Mode
// -> Type
// -> League
// -> LobbyId
// -> Address
// -> RconPassword
// -> LobbyPassword
//
func (s *Server) Setup() error {
	log.Println("[Server.Setup]: Setting up server -> [" + s.Address + "] from lobby [" + s.LobbyId.Hex() + "]")

	// connect to rcon
	var err error
	s.Rcon, err = TF2RconWrapper.NewTF2RconConnection(s.Address, s.RconPassword)

	if err != nil {
		return err
	}

	// changing server password
	passErr := s.Rcon.ChangeServerPassword(s.LobbyPassword)

	if passErr != nil {
		return passErr
	}

	// kick players
	log.Println("[Server.Setup]: Connected to server, getting players...")
	kickErr := s.KickAll()

	if kickErr != nil {
		return kickErr
	} else {
		log.Println("[Server.Setup]: Players kicked, running config!")
	}

	// run config
	config := NewServerConfig()
	config.League = s.League
	config.Type = s.Type
	config.Map = s.Map
	cfg, cfgErr := config.Get()

	if cfgErr == nil {
		config.Data = cfg
		configErr := s.ExecConfig(config)

		if configErr != nil {
			return configErr
		}
	} else {
		return cfgErr
	}

	// change map
	mapErr := s.Rcon.ChangeMap(s.Map)

	if mapErr != nil {
		return mapErr
	}

	// verify's timer
	s.Ticker.Ticker = time.NewTicker(10 * time.Second)
	s.Ticker.Quit = make(chan struct{})
	go func() {
		for {
			select {
			case <-s.Ticker.Ticker.C:
				s.Verify()
			case <-s.Ticker.Quit:
				s.Ticker.Ticker.Stop()
				return
			}
		}
	}()

	return nil
}

// runs each 10 sec
func (s *Server) Verify() {
	log.Println("[Server.Verify]: Verifing server -> [" + s.Address + "] from lobby [" + s.LobbyId.Hex() + "]")

	// check if all players in server are in lobby
	s.Players = s.Rcon.GetPlayers()
	for i := range s.Players {
		if s.Players[i].SteamID != "BOT" {
			commId, idErr := steamid.SteamIdToCommId(s.Players[i].SteamID)

			if idErr != nil {
				log.Printf("[Server.Verify]: ERROR -> %s", idErr)
			}

			isPlayerAllowed := s.IsPlayerAllowed(commId)

			if isPlayerAllowed == false {
				log.Println("[Server.Verify]: Kicking player not allowed -> Username [" +
					s.Players[i].Username + "] CommID [" + commId + "] SteamID [" + s.Players[i].SteamID + "] ")

				kickErr := s.Rcon.KickPlayer(s.Players[i], "[TeamPlay.TF]: You're not in this lobby...")

				if kickErr != nil {
					log.Printf("[Server.Verify]: ERROR -> %s", kickErr)
				}
			}
		}
	}
}

// check if the given commId is in the server
func (s *Server) IsPlayerInServer(playerCommId string) (bool, error) {
	for i := range s.Players {
		commId, idErr := steamid.SteamIdToCommId(s.Players[i].SteamID)

		if idErr != nil {
			return false, idErr
		}

		if playerCommId == commId {
			return true, nil
		}
	}

	return false, nil
}

// TODO: get end event from logs
// `World triggered "Game_Over"`
func (s *Server) End() {
	log.Println("[Server.End]: Ending server -> [" + s.Address + "] from lobby [" + s.LobbyId.Hex() + "]")
	// TODO: upload logs

	s.Rcon.Close()
	s.Ticker.Close()
}

func (s *Server) ExecConfig(config *ServerConfig) error {
	log.Println("[Server.ExecConfig]: Running config!")
	configErr := s.Rcon.ExecConfig(config.Data)

	if configErr != nil {
		log.Println("[Server.ExecConfig]: Error while trying to run config!")

		return configErr
	}

	return nil
}

func (s *Server) KickAll() error {
	log.Println("[Server.KickAll]: Kicking players...")
	s.Players = s.Rcon.GetPlayers()

	for i := range s.Players {
		kickErr := s.Rcon.KickPlayer(s.Players[i], "[TeamPlay.TF]: Setting up lobby...")

		if kickErr != nil {
			return kickErr
		}
	}

	return nil
}

func (s *Server) AllowPlayer(commId string) {
	s.AllowedPlayers = append(s.AllowedPlayers, TF2RconWrapper.Player{SteamID: commId})
}

func (s *Server) IsPlayerAllowed(commId string) bool {
	for i := range s.AllowedPlayers {
		if commId == s.AllowedPlayers[i].SteamID {
			return true
		}
	}

	return false
}
