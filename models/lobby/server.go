package lobby

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/TeamPlayTF/TF2RconWrapper"
)

type Server struct {
	Map  string // lobby map
	Name string // server name
	Rcon *TF2RconWrapper.TF2RconConnection

	League string
	Type   LobbyType // 9v9 6v6 4v4...

	Address string // server ip:port
	LobbyId int

	Players        []TF2RconWrapper.Player // current number of players in the server
	AllowedPlayers []TF2RconWrapper.Player

	Config     *ServerConfig // config that should run before the lobby starts
	WhiteList  string        // whitelist that should run before the lobby starts
	MaxPlayers int

	// timer that will verify()
	Ticker verifyTicker

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

type Map struct {
	Name   string
	Config string
	League string
	Mode   string
}

func NewServer() *Server {
	return new(Server)
}

// after create the server var, you should run this
//
// things that needs to be specified before run this:
// -> Map
// -> Mode
// -> League
// -> LobbyId
// -> Address
// -> RconPassword
// -> LobbyPassword
//
func (s *Server) Setup() error {
	fmt.Println("[Server.Setup]: Setting up server -> [" + s.Address + "] from lobby [#" + strconv.Itoa(s.LobbyId) + "]")

	// connect to rcon
	var err error
	s.Rcon, err = TF2RconWrapper.NewTF2RconConnection(s.Address, s.RconPassword)

	if err != nil {
		log.Fatal(err)
	}

	// changing server password
	passErr := s.ChangePassword(s.LobbyPassword)

	if passErr != nil {
		log.Fatal(passErr)
	}

	// kick players
	fmt.Println("[Server.Setup]: Connected to server, getting players...")
	kickErr := s.KickAll()

	if kickErr != nil {
		log.Fatal(kickErr)
	} else {
		fmt.Println("[Server.Setup]: Players kicked, running config!")
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
			log.Fatal(configErr)
		}
	} else {
		log.Fatal(cfgErr)
	}

	// change map
	mapErr := s.ChangeMap(s.Map)

	if mapErr != nil {
		log.Fatal(mapErr)
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
	fmt.Println("[Server.Verify]: Verifing server -> [" + s.Address + "] from lobby [#" + strconv.Itoa(s.LobbyId) + "]")

	// check if all players in server are in lobby
	s.Players = s.Rcon.GetPlayers()
	for i := range s.Players {
		// check if player is not in lobby but is in server
		// ignores BOT (SourceTV)
		if s.Players[i].SteamID != "BOT" && s.IsPlayerInLobby(s.Players[i].SteamID) == false {
			fmt.Println("[Server.Verify]: Kicking player not allowed -> Username [" +
				s.Players[i].Username + "] SteamID [" + s.Players[i].SteamID + "]")

			kickErr := s.Rcon.KickPlayer(s.Players[i], "[TeamPlay.TF]: You're not in this lobby...")

			if kickErr != nil {
				log.Fatal(kickErr)
			}
		}
	}

}

// check if the given steamId is in the server
func (s *Server) IsPlayerInServer(steamId string) bool {
	for i := range s.Players {
		if steamId == s.Players[i].SteamID {
			return true
		}
	}

	return false
}

// check if the given steamId is in the allowedPlayers list
func (s *Server) IsPlayerInLobby(steamId string) bool {
	// SourceTV
	if steamId == "BOT" {
		return false
	}

	for i := range s.AllowedPlayers {
		if steamId == s.AllowedPlayers[i].SteamID {
			return true
		}
	}

	return false
}

// TODO: get end event from logs
// `World triggered "Game_Over"`
func (s *Server) End() {
	fmt.Println("[Server.End]: Ending server -> [" + s.Address + "] from lobby [" + strconv.Itoa(s.LobbyId) + "]")
	// TODO: upload logs

	s.Rcon.Close()
	s.Ticker.Close()
}

func (s *Server) ExecConfig(config *ServerConfig) error {
	fmt.Println("[Server.ExecConfig]: Running config!")
	configErr := s.Rcon.ExecConfig(config.Data)

	if configErr != nil {
		fmt.Println("[Server.ExecConfig]: Error while trying to run config!")

		return configErr
	}

	return nil
}

func (s *Server) KickAll() error {
	fmt.Println("[Server.KickAll]: Kicking players...")
	s.Players = s.Rcon.GetPlayers()

	for i := range s.Players {
		kickErr := s.Rcon.KickPlayer(s.Players[i], "[TeamPlay.TF]: Setting up lobby...")

		if kickErr != nil {
			return kickErr
		}
	}

	return nil
}

func (s *Server) ChangeMap(newMap string) error {
	fmt.Println("[Server.ChangeMap]: Changing [" + s.Address + "]'s map to [" + newMap + "]...")
	_, queryErr := s.Rcon.Query("changelevel " + newMap)

	return queryErr
}

func (s *Server) ChangePassword(newPassword string) error {
	// reset password if length is 0
	// command would be: sv_password ""
	if newPassword == "" {
		newPassword = `""`
	}

	fmt.Println("[Server.ChangePassword]: Changing [" + strconv.Itoa(s.LobbyId) + " @ " + s.Address + "]'s password...")
	_, queryErr := s.Rcon.Query("sv_password " + newPassword)

	if queryErr == nil {
		fmt.Println("[Server.ChangePassword]: Changed [" + strconv.Itoa(s.LobbyId) + " @ " + s.Address + "]'s password!")
	}

	return queryErr
}

func (s *Server) AllowPlayer(steamId string) {
	s.AllowedPlayers = append(s.AllowedPlayers, TF2RconWrapper.Player{SteamID: steamId})
}
