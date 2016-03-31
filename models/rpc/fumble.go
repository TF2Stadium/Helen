package rpc

import "github.com/Sirupsen/logrus"

func FumbleLobbyCreated(lobbyID uint) error {
	if *fumbleDisabled {
		return nil
	}

	err := fumble.Call("Fumble.CreateLobby", lobbyID, &struct{}{})

	if err != nil {
		logrus.Error(err)
		return err
	}

	return nil
}

func FumbleLobbyEnded(lobbyID uint) {
	if *fumbleDisabled {
		return
	}

	err := fumble.Call("Fumble.EndLobby", lobbyID, &struct{}{})
	if err != nil {
		logrus.Error(err)
	}
}
