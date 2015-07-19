package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/TF2Stadium/Server/config"
	"github.com/TF2Stadium/Server/controllers/controllerhelpers"
	"github.com/TF2Stadium/Server/database"
	"github.com/TF2Stadium/Server/models"
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
		log.Print(err)
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
		log.Println(err)
		return
	}

	parts := strings.Split(id, "/")
	steamid := parts[len(parts)-1]

	session, _ := controllerhelpers.GetSessionHTTP(r)
	session.Values["steam_id"] = steamid

	player := &models.Player{}
	err = database.DB.Where("steam_id = ?", steamid).First(player).Error

	if err == gorm.RecordNotFound {
		player = models.NewPlayer(steamid)
		database.DB.Create(player)
	} else if err != nil {
		log.Println("steamLogin.go:60 ", err)
	}

	session.Values["id"] = fmt.Sprint(player.ID)

	log.Println(session)

	err = session.Save(r, w)

	http.Redirect(w, r, config.Constants.LoginRedirectPath, 303)
}
