// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package login

import (
	"net/http"
	"regexp"
	"strconv"
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

var (
	nonceStore = &openid.SimpleNonceStore{
		Store: make(map[string][]*openid.Nonce)}
	discoveryCache = &openid.SimpleDiscoveryCache{}
)

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

func LogoutSession(w http.ResponseWriter, r *http.Request) {
	session, _ := controllerhelpers.GetSessionHTTP(r)
	session.Options = &sessions.Options{MaxAge: -1}
	session.Save(r, w)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	LogoutSession(w, r)

	http.Redirect(w, r, config.Constants.LoginRedirectPath, 303)
}

func setSession(w http.ResponseWriter, r *http.Request, steamid string) error {
	session, _ := controllerhelpers.GetSessionHTTP(r)
	session.Values["steam_id"] = steamid

	player := &models.Player{}
	err := database.DB.Where("steam_id = ?", steamid).First(player).Error

	var playErr error
	if err == gorm.RecordNotFound {
		// Successful first-time login
		player, playErr = models.NewPlayer(steamid)

		if playErr != nil {
			return playErr
		}

		database.DB.Create(player)
	} else if err == nil {
		// Successful repeat login
		err = player.UpdatePlayerInfo()
		if err == nil {
			player.Save()
		} else {
			return err
		}
	} else if err != nil {
		// Failed login
		return err
	}

	session.Values["id"] = strconv.FormatUint(uint64(player.ID), 10)
	session.Values["role"] = player.Role

	session.Options.Domain = config.Constants.CookieDomain
	err = session.Save(r, w)
	return err
}

func LoginCallbackHandler(w http.ResponseWriter, r *http.Request) {
	regex := regexp.MustCompile(`http://steamcommunity.com/openid/id/(\d+)`)

	fullURL := config.Constants.Domain + r.URL.String()
	idURL, err := openid.Verify(fullURL, discoveryCache, nonceStore)
	if err != nil {
		helpers.Logger.Warning("%s", err.Error())
		return
	}

	parts := regex.FindStringSubmatch(idURL)
	if len(parts) != 2 {
		http.Error(w, "Steam Authentication failed, please try again.", 500)
		return
	}

	steamid := parts[1]

	if config.Constants.SteamIDWhitelist != "" &&
		!controllerhelpers.IsSteamIDWhitelisted(steamid) {
		//Use a more user-friendly message later
		http.Error(w, "Sorry, you're not in the closed alpha.", 403)
		return
	}
	err = setSession(w, r, steamid)
	if err != nil {
		session, _ := controllerhelpers.GetSessionHTTP(r)
		session.Options = &sessions.Options{MaxAge: -1}
		session.Save(r, w)

		helpers.Logger.Error(err.Error())
		http.Error(w, "Internal Server Error.", 500)
		return
	}
	http.Redirect(w, r, config.Constants.LoginRedirectPath, 303)
}
