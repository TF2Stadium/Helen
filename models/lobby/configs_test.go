package lobby

import (
	"log"
	"testing"

	"github.com/TeamPlayTF/Server/config"
	"github.com/TeamPlayTF/Server/models/lobby"
)

func TestInitConfigs(t *testing.T) {
	config.SetupConstants()
	lobby.InitConfigs()
}

func TestUgcHighlander(t *testing.T) {
	config := lobby.NewServerConfig()
	config.League = "ugc"
	config.Type = lobby.LobbyTypeHighlander
	config.Map = "cp_process_final"
	_, cfgErr := config.Get()

	if cfgErr != nil {
		log.Fatal(cfgErr)
	}
}

func TestUgcSixes(t *testing.T) {
	config := lobby.NewServerConfig()
	config.League = "ugc"
	config.Type = lobby.LobbyTypeSixes
	config.Map = "cp_process_final"
	_, cfgErr := config.Get()

	if cfgErr != nil {
		log.Fatal(cfgErr)
	}
}

func TestEtf2lSixes(t *testing.T) {
	config := lobby.NewServerConfig()
	config.League = "etf2l"
	config.Type = lobby.LobbyTypeSixes
	config.Map = "cp_process_final"
	_, cfgErr := config.Get()

	if cfgErr != nil {
		log.Fatal(cfgErr)
	}
}

func TestEtf2lHighlander(t *testing.T) {
	config := lobby.NewServerConfig()
	config.League = "etf2l"
	config.Type = lobby.LobbyTypeHighlander
	config.Map = "cp_process_final"
	_, cfgErr := config.Get()

	if cfgErr != nil {
		log.Fatal(cfgErr)
	}
}
