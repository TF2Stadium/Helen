// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package helpers

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	dg "github.com/bwmarrin/discordgo"
)

var (
	Discord *dg.Session
	emojis = make(map[string]string)
	channels = make(map[string]*dg.Channel)
)

func DiscordSendToChannel(channelName string, msg string) {
	if channel, ok := channels[channelName]; ok {
		Discord.ChannelMessageSend(channel.ID, msg)
	}
}

func DiscordEmoji(emoji string) string {
	code, customEmojiExists := emojis[emoji]
	if !customEmojiExists {
		code = fmt.Sprintf(":%s:", emoji)
	}
	return code
}

func init() {
	token := config.Constants.DiscordToken
	guildId := config.Constants.DiscordGuildId
	if token == "" || guildId == "" {
		return
	}

	var err error
	Discord, err = dg.New(token)
	if err != nil {
		logrus.Fatal(err)
		return
	}

	guild, err := Discord.Guild(guildId)
	if err != nil {
		Discord = nil
		logrus.Fatal(err)
		return
	}

	rawChannels, err := Discord.GuildChannels(guildId)
	if err != nil {
		Discord = nil
		logrus.Fatal(err)
		return
	}

	for _, emoji := range guild.Emojis {
		emojis[emoji.Name] = fmt.Sprintf("<:%s:%s>", emoji.Name, emoji.ID)
	}
	for _, channel := range rawChannels {
		channels[channel.Name] = channel
	}
	logrus.Infof("Discord: Loaded %d channels, %d emojis", len(rawChannels), len(guild.Emojis))
}
