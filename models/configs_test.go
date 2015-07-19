package models

import (
	"testing"

	"github.com/TF2Stadium/Server/config"
	"github.com/stretchr/testify/assert"
)

func TestInitConfigs(t *testing.T) {
	config.SetupConstants()
	InitServerConfigs()
}

func TestUgcHighlander(t *testing.T) {
	config := NewServerConfig()
	config.League = LeagueUgc
	config.Type = LobbyTypeHighlander
	config.Map = "pl_upward"
	cfg, cfgErr := config.Get()

	assert.Nil(t, cfgErr, "cfg error")
	assert.NotEmpty(t, cfg, "cfg shouldn't be empty")
}

func TestUgcSixes(t *testing.T) {
	config := NewServerConfig()
	config.League = LeagueUgc
	config.Type = LobbyTypeSixes
	config.Map = "cp_badlands"
	cfg, cfgErr := config.Get()

	assert.Nil(t, cfgErr, "cfg error")
	assert.NotEmpty(t, cfg, "cfg shouldn't be empty")
}

func TestEtf2lSixes(t *testing.T) {
	config := NewServerConfig()
	config.League = LeagueEtf2l
	config.Type = LobbyTypeSixes
	config.Map = "cp_gullywash_final1"
	cfg, cfgErr := config.Get()

	assert.Nil(t, cfgErr, "cfg error")
	assert.NotEmpty(t, cfg, "cfg shouldn't be empty")
}

func TestEtf2lHighlander(t *testing.T) {
	config := NewServerConfig()
	config.League = LeagueEtf2l
	config.Type = LobbyTypeHighlander
	config.Map = "pl_badwater"
	cfg, cfgErr := config.Get()

	assert.Nil(t, cfgErr, "cfg error")
	assert.NotEmpty(t, cfg, "cfg shouldn't be empty")
}
