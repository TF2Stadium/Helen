package models

import (
	"fmt"
	"log"
	"time"

	"github.com/TF2Stadium/PlayerStatsScraper/steamid"
	"github.com/TF2Stadium/Server/config"
	"github.com/TF2Stadium/Server/helpers"
	"github.com/TF2Stadium/TF2RconWrapper"
)

var LobbyServerMap = make(map[uint]*Server)
var LobbyServerSettingUp = make(map[uint]time.Time)

type ServerRecord struct {
	ID           uint
	Host         string
	RconPassword string
}

type Server struct {
	Map  string // lobby map
	Name string // server name

	League League
	Type   LobbyType // 9v9 6v6 4v4...

	LobbyId uint

	Players        []TF2RconWrapper.Player // current number of players in the server
	AllowedPlayers map[string]bool

	Config *ServerConfig // config that should run before the lobby starts
	Ticker verifyTicker  // timer that will verify()

	//ChatListener  *TF2RconWrapper.RconChatListener

	Rcon          *TF2RconWrapper.TF2RconConnection
	Info          ServerRecord
	LobbyPassword string // will store the new server password from the lobby
}

// timer used in verify()
type verifyTicker struct {
	Ticker *time.Ticker
	Quit   chan bool
}

func (t *verifyTicker) Close() {
	t.Quit <- true
}

func NewServer() *Server {
	s := &Server{}
	s.AllowedPlayers = make(map[string]bool)

	return s
}

// after create the server var, you should run this
//
// things that needs to be specified before run this:
// -> Map
// -> Type
// -> League
// -> Info
//

func (s *Server) VerifyInfo() error {
	if config.Constants.ServerMockUp {
		return nil
	}

	var err error
	s.Rcon, err = TF2RconWrapper.NewTF2RconConnection(s.Info.Host, s.Info.RconPassword)

	if err != nil {
		return helpers.NewTPError(err.Error(), -1)
	}
	return nil
}

func (s *Server) SetupObject() error {
	// If the ticker is initialized, the verifier is running
	if s.Ticker.Quit != nil {
		return nil
	}
	s.Ticker.Ticker = time.NewTicker(10 * time.Second)
	s.Ticker.Quit = make(chan bool)
	go func() {
		for {
			select {
			case <-s.Ticker.Ticker.C:
				s.Verify()
			case <-s.Ticker.Quit:
				log.Println("Stopping verifier")
				s.Ticker.Ticker.Stop()
				return
			}
		}
	}()

	return nil
}

func (s *Server) Setup() error {
	if config.Constants.ServerMockUp {
		return nil
	}

	helpers.Logger.Debug("[Server.Setup]: Setting up server -> [" + s.Info.Host + "] from lobby [" + fmt.Sprint(s.LobbyId) + "]")

	// connect to rcon if not connected before
	if s.Rcon == nil {
		var err error
		s.Rcon, err = TF2RconWrapper.NewTF2RconConnection(s.Info.Host, s.Info.RconPassword)

		if err != nil {
			return err
		}
	}

	// changing server password
	passErr := s.Rcon.ChangeServerPassword(s.LobbyPassword)

	if passErr != nil {
		return passErr
	}

	// kick players
	helpers.Logger.Debug("[Server.Setup]: Connected to server, getting players...")
	kickErr := s.KickAll()

	if kickErr != nil {
		return kickErr
	} else {
		helpers.Logger.Debug("[Server.Setup]: Players kicked, running config!")
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

	return nil
}

// runs each 10 sec
func (s *Server) Verify() {
	if config.Constants.ServerMockUp || s.Rcon == nil {
		return
	}
	helpers.Logger.Debug("[Server.Verify]: Verifing server -> [" + s.Info.Host + "] from lobby [" + fmt.Sprint(s.LobbyId) + "]")

	// check if all players in server are in lobby
	var err error
	s.Players, err = s.Rcon.GetPlayers()

	for err != nil {
		time.Sleep(time.Second)
		helpers.Logger.Warning("Failed to get players in server %s: %s", s.LobbyId, err.Error())
		s.Players, err = s.Rcon.GetPlayers()
	}

	for i := range s.Players {
		if s.Players[i].SteamID != "BOT" {
			commId, idErr := steamid.SteamIdToCommId(s.Players[i].SteamID)

			if idErr != nil {
				helpers.Logger.Debug("[Server.Verify]: ERROR -> %s", idErr)
			}

			isPlayerAllowed := s.IsPlayerAllowed(commId)

			if isPlayerAllowed == false {
				helpers.Logger.Debug("[Server.Verify]: Kicking player not allowed -> Username [" +
					s.Players[i].Username + "] CommID [" + commId + "] SteamID [" + s.Players[i].SteamID + "] ")

				kickErr := s.Rcon.KickPlayer(s.Players[i], "[tf2stadium.com]: You're not in this lobby...")

				if kickErr != nil {
					helpers.Logger.Debug("[Server.Verify]: ERROR -> %s", kickErr)
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
	if config.Constants.ServerMockUp {
		return
	}

	helpers.Logger.Debug("[Server.End]: Ending server -> [" + s.Info.Host + "] from lobby [" + fmt.Sprint(s.LobbyId) + "]")
	// TODO: upload logs

	s.Rcon.Close()
	s.Ticker.Close()
}

func (s *Server) ExecConfig(config *ServerConfig) error {
	helpers.Logger.Debug("[Server.ExecConfig]: Running config!")
	configErr := s.Rcon.ExecConfig(config.Data)

	if configErr != nil {
		helpers.Logger.Debug("[Server.ExecConfig]: Error while trying to run config!")

		return configErr
	}

	return nil
}

func (s *Server) KickAll() error {
	helpers.Logger.Debug("[Server.KickAll]: Kicking players...")
	var err error
	s.Players, err = s.Rcon.GetPlayers()

	for err != nil {
		time.Sleep(time.Second)
		helpers.Logger.Warning("Failed to get players in server %s: %s", s.LobbyId, err.Error())
		s.Players, err = s.Rcon.GetPlayers()
	}

	for i := range s.Players {
		kickErr := s.Rcon.KickPlayer(s.Players[i], "[tf2stadium.com]: Setting up lobby...")

		if kickErr != nil {
			return kickErr
		}
	}

	return nil
}

func (s *Server) SetAllowedPlayers(commIds []string) {
	s.AllowedPlayers = make(map[string]bool)

	for _, commId := range commIds {
		s.AllowedPlayers[commId] = true
	}
}

func (s *Server) AllowPlayer(commId string) {
	s.AllowedPlayers[commId] = true
}

func (s *Server) DisallowPlayer(commId string) {
	if s.IsPlayerAllowed(commId) {
		delete(s.AllowedPlayers, commId)
	}
}

func (s *Server) IsPlayerAllowed(commId string) bool {
	if _, ok := s.AllowedPlayers[commId]; ok {
		return true
	}

	return false
}
