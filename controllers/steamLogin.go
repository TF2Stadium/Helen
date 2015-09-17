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

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := controllerhelpers.GetSessionHTTP(r)
	session.Options = &sessions.Options{MaxAge: -1}
	session.Save(r, w)

	http.Redirect(w, r, "/", 303)
}

func LoginCallbackHandler(w http.ResponseWriter, r *http.Request) {
	fullURL := config.Constants.Domain + r.URL.String()
	id, err := openid.Verify(fullURL, discoveryCache, nonceStore)
	if err != nil {
		helpers.Logger.Debug(err.Error())
		return
	}

	parts := strings.Split(id, "/")
	steamid := parts[len(parts)-1]

	session, _ := controllerhelpers.GetSessionHTTP(r)
	session.Values["steam_id"] = steamid

	player := &models.Player{}
	var playErr error
	err = database.DB.Where("steam_id = ?", steamid).First(player).Error

	if err == gorm.RecordNotFound {
		player, playErr = models.NewPlayer(steamid)

		if playErr != nil {
			helpers.Logger.Debug(playErr.Error())
		}

		database.DB.Create(player)
	} else if err != nil {
		helpers.Logger.Debug("steamLogin.go:60 ", err)
	}

	session.Values["id"] = fmt.Sprint(player.ID)
	session.Values["role"] = player.Role

	session.Options.Domain = config.Constants.CookieDomain
	err = session.Save(r, w)

	http.Redirect(w, r, config.Constants.LoginRedirectPath, 303)
}
