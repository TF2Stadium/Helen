package controllers

import (
	"net/http"
	"time"

	"github.com/TF2Stadium/Helen/config"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
)

func ResetMumblePassword(w http.ResponseWriter, r *http.Request) {
	token, err := chelpers.GetToken(r)
	if err != nil {
		http.Error(w, "You aren't logged in.", http.StatusForbidden)
		return
	}

	player := chelpers.GetPlayer(token)
	player.MumbleAuthkey = player.GenAuthKey()
	player.Save()

	newToken := chelpers.NewToken(player)
	cookie := &http.Cookie{
		Name:    "auth-jwt",
		Value:   newToken,
		Path:    "/",
		Domain:  config.Constants.CookieDomain,
		Expires: time.Now().Add(30 * 24 * time.Hour),
	}
	http.SetCookie(w, cookie)

	referer, ok := r.Header["Referer"]
	if ok {
		http.Redirect(w, r, referer[0], 303)
		return
	}

	http.Redirect(w, r, config.Constants.LoginRedirectPath, 303)
}
