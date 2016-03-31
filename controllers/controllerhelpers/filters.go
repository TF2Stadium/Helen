// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllerhelpers

import (
	"encoding/xml"
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/helpers/authority"
	"github.com/TF2Stadium/Helen/models/player"
	"github.com/TF2Stadium/wsevent"
)

var (
	whitelistLock    = new(sync.RWMutex)
	whitelistSteamID map[string]bool
)

func WhitelistListener() {
	ticker := time.NewTicker(time.Minute * 1)
	for {
		client := http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(config.Constants.SteamIDWhitelist)

		if err != nil {
			logrus.Error(err.Error())
			continue
		}

		var groupXML struct {
			//XMLName xml.Name `xml:"memberList"`
			//GroupID uint64   `xml:"groupID64"`
			Members []string `xml:"members>steamID64"`
		}

		dec := xml.NewDecoder(resp.Body)
		err = dec.Decode(&groupXML)
		if err != nil {
			logrus.Error(err)
			continue
		}

		whitelistLock.Lock()
		whitelistSteamID = make(map[string]bool)

		for _, steamID := range groupXML.Members {
			//_, ok := whitelistSteamID[steamID]
			//logrus.Info("Whitelisting SteamID %s", steamID)
			whitelistSteamID[steamID] = true
		}
		whitelistLock.Unlock()
		<-ticker.C
	}
}

func IsSteamIDWhitelisted(steamid string) bool {
	whitelistLock.RLock()
	defer whitelistLock.RUnlock()
	whitelisted, exists := whitelistSteamID[steamid]

	return whitelisted && exists
}

// shitlord
func CheckPrivilege(so *wsevent.Client, action authority.AuthAction) error {
	//Checks if the client has the neccesary authority to perform action
	player, _ := player.GetPlayerBySteamID(so.Token.Claims["steam_id"].(string))
	if !player.Role.Can(action) {
		return errors.New("You are not authorized to perform this action")
	}
	return nil
}

func FilterHTTPRequest(action authority.AuthAction, f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		token, err := GetToken(r)
		if err != nil {
			http.Error(w, "You're not logged in, or your JWT cookie is invalid.", http.StatusBadRequest)
			return
		}

		role, _ := strconv.Atoi(token.Claims["role"].(string))

		if !(authority.AuthRole(role).Can(action)) {
			http.Error(w, "Not authorized", 403)
			return
		}

		f(w, r)
	}
}
