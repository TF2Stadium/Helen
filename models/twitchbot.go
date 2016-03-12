package models

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
