package rpc

func TwitchBotJoin(channel string) {
	if *twitchbotDisabled {
		return
	}
	twitchbot.Call("TwitchBot.Join", channel, &struct{}{})
}

func TwitchBotLeave(channel string) {
	if *twitchbotDisabled {
		return
	}
	twitchbot.Call("TwitchBot.Leave", channel, &struct{}{})
}

func TwitchBotAnnouce(channel string, lobbyid uint) {
	if *twitchbotDisabled {
		return
	}

	twitchbot.Go("TwitchBot.Announce", struct {
		Channel string
		LobbyID uint
	}{channel, lobbyid}, &struct{}{}, nil)
}
