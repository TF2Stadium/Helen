package models

import (
	"time"

	db "github.com/TF2Stadium/Helen/database"
)

// Requirement stores a requirement for a particular slot in a lobby
type Requirement struct {
	ID      uint `json:"-"`
	LobbyID uint `json:"-"`

	Slot int `json:"-"` // if -1, applies to all slots

	Hours       int     `json:"hours"`       // minimum hours needed
	Lobbies     int     `json:"lobbies"`     // minimum lobbies played
	Reliability float64 `json:"reliability"` // minimum reliability needed
}

func NewRequirement(lobbyID uint, slot int, hours int, lobbies int) *Requirement {
	r := &Requirement{
		LobbyID: lobbyID,
		Slot:    slot,
		Hours:   hours,
		Lobbies: lobbies}
	db.DB.Save(r)

	return r
}

func (r *Requirement) Save() { db.DB.Save(r) }

//GetSlotRequirement returns the slot requirement for the lobby lobby
func (lobby *Lobby) GetSlotRequirement(slot int) (*Requirement, error) {
	req := &Requirement{}
	err := db.DB.Table("requirements").Where("lobby_id = ? AND slot = ?", lobby.ID, slot).First(req).Error

	return req, err
}

//HasSlotRequirement returns true if the given slot in the lobby has a requirement
func (lobby *Lobby) HasSlotRequirement(slot int) bool {
	var count int
	db.DB.Table("requirements").Where("lobby_id = ? AND slot = ?", lobby.ID, slot).Count(&count)
	return count != 0
}

//HasRequirements returns true if the given slot has a requirement (either general or slot-only)
func (lobby *Lobby) HasRequirements(slot int) bool {
	return lobby.HasSlotRequirement(slot)
}

//FitsRequirements checks if the player fits the requirement to be added to the given slot in the lobby
func (l *Lobby) FitsRequirements(player *Player, slot int) (bool, error) {
	//BUG(vibhavp): FitsRequirements doesn't check reliability
	var req *Requirement

	slotReq, err := l.GetSlotRequirement(slot)
	if err == nil {
		req = slotReq
	}

	db.DB.Preload("Stats").First(player, player.ID)

	if time.Since(player.ProfileUpdatedAt) < time.Hour*time.Duration(req.Hours-player.GameHours) {
		//update player info only if the number of hours needed > the number of hours
		//passed since player info was last updated
		player.UpdatePlayerInfo()
		player.Save()
	}

	if player.GameHours < req.Hours {
		return false, ErrReqHours
	}

	if player.Stats.TotalLobbies() < req.Lobbies {
		return false, ErrReqLobbies
	}

	return true, nil
}
