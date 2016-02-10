package login

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	db "github.com/TF2Stadium/Helen/database"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"golang.org/x/net/xsrftoken"
)

type reply struct {
	AccessToken string   `json:"access_token"`
	Scope       []string `json:"scope"`
}

type userInfo struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

func TwitchLogin(w http.ResponseWriter, r *http.Request) {
	session, err := controllerhelpers.GetSessionHTTP(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	steamID, ok := session.Values["steam_id"]
	if !ok {
		http.Error(w, "You are not logged in.", http.StatusUnauthorized)
		return
	}

	player, _ := models.GetPlayerBySteamID(steamID.(string))

	loginURL := url.URL{
		Scheme: "https",
		Host:   "api.twitch.tv",
		Path:   "kraken/oauth2/authorize",
	}

	twitchRedirectURL := "http://" + config.Constants.ListenAddress + "/" + "twitchAuth"

	values := loginURL.Query()
	values.Set("response_type", "code")
	values.Set("client_id", config.Constants.TwitchClientID)
	values.Set("redirect_uri", twitchRedirectURL)
	values.Set("scope", "channel_check_subscription user_subscriptions channel_subscriptions user_read")
	values.Set("state", xsrftoken.Generate(config.Constants.CookieStoreSecret, player.SteamID, "GET"))
	loginURL.RawQuery = values.Encode()

	http.Redirect(w, r, loginURL.String(), http.StatusTemporaryRedirect)
}

func TwitchAuth(w http.ResponseWriter, r *http.Request) {
	session, err := controllerhelpers.GetSessionHTTP(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	steamID, ok := session.Values["steam_id"]
	if !ok {
		http.Error(w, "You are not logged in.", http.StatusUnauthorized)
	}

	player, _ := models.GetPlayerBySteamID(steamID.(string))

	values := r.URL.Query()
	code := values.Get("code")
	if code == "" {
		http.Error(w, "No code given", http.StatusBadRequest)
		return
	}

	state := values.Get("state")
	if state == "" || !xsrftoken.Valid(state, config.Constants.CookieStoreSecret, player.SteamID, "GET") {
		http.Error(w, "Missing or Invalid XSRF token", http.StatusBadRequest)
		return
	}

	twitchRedirectURL := "http://" + config.Constants.ListenAddress + "/" + "twitchAuth"

	// successful login, try getting access token now
	tokenURL := url.URL{
		Scheme: "https",
		Host:   "api.twitch.tv",
		Path:   "kraken/oauth2/token",
	}
	values = tokenURL.Query()
	values.Set("client_id", config.Constants.TwitchClientID)
	values.Set("client_secret", config.Constants.TwitchClientSecret)
	values.Set("grant_type", "authorization_code")
	values.Set("redirect_uri", twitchRedirectURL)
	values.Set("code", code)
	values.Set("state", state)

	req, err := http.NewRequest("POST", tokenURL.String(), strings.NewReader(values.Encode()))
	if err != nil {
		logrus.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp, err := helpers.HTTPClient.Do(req)
	if err != nil {
		logrus.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	reply := reply{}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&reply)
	if err != nil {
		logrus.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	info, err := getUserInfo(reply.AccessToken)
	if err != nil {
		logrus.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	player.TwitchName = info.Name
	player.TwitchAccessToken = reply.AccessToken
	player.Save()

	http.Redirect(w, r, config.Constants.LoginRedirectPath, http.StatusTemporaryRedirect)
}

func getUserInfo(token string) (*userInfo, error) {
	req, _ := http.NewRequest("GET", "https://api.twitch.tv/kraken/user", nil)
	req.Header.Add("Accept", "application/vnd.twitchtv.v3+json")
	req.Header.Add("Authorization", "OAuth "+token)

	resp, err := helpers.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	info := &userInfo{}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(info)
	return info, err
}

func TwitchLogout(w http.ResponseWriter, r *http.Request) {
	session, err := controllerhelpers.GetSessionHTTP(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	steamID, ok := session.Values["steam_id"]
	if !ok {
		http.Error(w, "You are not logged in.", http.StatusUnauthorized)
	}

	db.DB.Table("players").Where("steam_id = ?", steamID).UpdateColumn(map[string]interface{}{
		"twitch_access_token": "",
		"twitch_name":         ""})
	http.Redirect(w, r, config.Constants.LoginRedirectPath, http.StatusTemporaryRedirect)
}
