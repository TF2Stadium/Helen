// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	"github.com/yohcop/openid-go"
)

var nonceStore = &openid.SimpleNonceStore{
	Store: make(map[string][]*openid.Nonce)}
var discoveryCache = &openid.SimpleDiscoveryCache{}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if url, err := openid.RedirectURL("http://steamcommunity.com/openid",
		config.Constants.Domain+"/openidcallback",
		config.Constants.OpenIDRealm); err == nil {
		http.Redirect(w, r, url, 303)
	} else {
		helpers.Logger.Debug(err.Error())
	}
}

func MockLoginHandler(w http.ResponseWriter, r *http.Request) {
	steamid := r.URL.Path[strings.Index(r.URL.Path, "Login/")+6:]
	setSession(w, r, steamid)
	http.Redirect(w, r, config.Constants.LoginRedirectPath, 303)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := controllerhelpers.GetSessionHTTP(r)
	session.Options = &sessions.Options{MaxAge: -1}
	session.Save(r, w)

	http.Redirect(w, r, config.Constants.LoginRedirectPath, 303)
}

func setSession(w http.ResponseWriter, r *http.Request, steamid string) {
	session, _ := controllerhelpers.GetSessionHTTP(r)
	session.Values["steam_id"] = steamid

	player := &models.Player{}
	err := database.DB.Where("steam_id = ?", steamid).First(player).Error

	var playErr error
	if err == gorm.RecordNotFound {
		// Successful first-time login
		player, playErr = models.NewPlayer(steamid)

		if playErr != nil {
			helpers.Logger.Warning(playErr.Error())
		}

		database.DB.Create(player)
	} else if err == nil {
		// Successful repeat login
		err = player.UpdatePlayerInfo()
		if err == nil {
			database.DB.Save(player)
		} else {
			helpers.Logger.Warning("Error updating player ", err)
		}
	} else if err != nil {
		// Failed login
		helpers.Logger.Warning("%s", err)
	}

	session.Values["id"] = fmt.Sprint(player.ID)
	session.Values["role"] = player.Role

	session.Options.Domain = config.Constants.CookieDomain
	err = session.Save(r, w)
}

func LoginCallbackHandler(w http.ResponseWriter, r *http.Request) {
	fullURL := config.Constants.Domain + r.URL.String()
	id, err := openid.Verify(fullURL, discoveryCache, nonceStore)
	if err != nil {
		helpers.Logger.Warning(err.Error())
		return
	}

	parts := strings.Split(id, "/")
	steamid := parts[len(parts)-1]
	if config.Constants.SteamIDWhitelist != "" && !controllerhelpers.IsSteamIDWhitelisted(steamid) {
		//Use a more user-friendly message later
		http.Error(w, "Sorry, you're not in the closed alpha.", 403)
		return
	}
	setSession(w, r, steamid)
	http.Redirect(w, r, config.Constants.LoginRedirectPath, 303)
}
