package models

import "github.com/Sirupsen/logrus"

func fumbleLobbyCreated(lobbyID uint) error {
	if !*fumbleEnabled {
		return nil
	}

	err := fumble.Call("Fumble.CreateLobby", lobbyID, nil)

	if err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}

func fumbleLobbyEnded(lobbyID uint) {
	if !*fumbleEnabled {
		return
	}

	err := fumble.Call("Fumble.EndLobby", lobbyID, nil)
	if err != nil {
		logrus.Error(err)
	}
}
