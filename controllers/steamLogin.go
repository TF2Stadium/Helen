package controllers

import (
	"log"
	"net/http"
	"strings"

	"gopkg.in/mgo.v2/bson"

	"github.com/TeamPlayTF/Server/config"
	"github.com/TeamPlayTF/Server/database"
	"github.com/TeamPlayTF/Server/models"
	"github.com/gorilla/sessions"
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
	session := getDefaultSession(r)
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

	session := getDefaultSession(r)
	session.Values["steamid"] = steamid

	player := &models.Player{}
	err = database.GetPlayersCollection().Find(bson.M{"steamid": steamid}).One(&player)

	if err != nil {
		player := models.NewPlayer(steamid)
		player.Save()
	}

	session.Values["id"] = player.Id.String()

	err = session.Save(r, w)

	http.Redirect(w, r, config.Constants.LoginRedirectPath, 303)
}
