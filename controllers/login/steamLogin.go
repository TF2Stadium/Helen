// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package login

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/models"
	"github.com/yohcop/openid-go"
)

var (
	nonceStore     = openid.NewSimpleNonceStore()
	discoveryCache = openid.NewSimpleDiscoveryCache()
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if url, err := openid.RedirectURL("http://steamcommunity.com/openid",
		config.Constants.PublicAddress+"/openidcallback",
		config.Constants.OpenIDRealm); err == nil {
		http.Redirect(w, r, url, 303)
	} else {
		logrus.Error(err)
	}
}

func MockLoginHandler(w http.ResponseWriter, r *http.Request) {
	steamid := r.URL.Path[strings.Index(r.URL.Path, "Login/")+6:]

	var player *models.Player
	var err error

	player, tperr := models.GetPlayerBySteamID(steamid)
	if tperr != nil {
		player, err = models.NewPlayer(steamid)
		if err != nil {
			logrus.Error(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		database.DB.Create(player)
	}

	player.UpdatePlayerInfo()

	http.Redirect(w, r, config.Constants.LoginRedirectPath, 303)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("auth-jwt")
	if err != nil { //user wasn't even logged in ಠ_ಠ
		return
	}

	cookie.MaxAge = -1
	cookie.Expires = time.Time{}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, config.Constants.PublicAddress, 303)
}

func LoginCallbackHandler(w http.ResponseWriter, r *http.Request) {
	regex := regexp.MustCompile(`http://steamcommunity.com/openid/id/(\d+)`)

	fullURL := config.Constants.PublicAddress + r.URL.String()
	idURL, err := openid.Verify(fullURL, discoveryCache, nonceStore)
	if err != nil {
		http.Error(w, "Steam Authentication failed, please try again.", 500)
		logrus.Error(err)
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

	var player *models.Player
	player, tperr := models.GetPlayerBySteamID(steamid)
	if tperr != nil {
		player, err = models.NewPlayer(steamid)
		if err != nil {
			logrus.Error(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		database.DB.Create(player)
	}

	player.UpdatePlayerInfo()
	key := controllerhelpers.NewToken(player.ID, steamid, player.Role)
	cookie := &http.Cookie{
		Name:    "auth-jwt",
		Value:   key,
		Path:    "/",
		Domain:  config.Constants.CookieDomain,
		Expires: time.Now().Add(30 * 24 * time.Hour),
		//Secure: true,
	}

	http.SetCookie(w, cookie)
	http.Redirect(w, r, config.Constants.LoginRedirectPath, 303)
}
