package login

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models/player"
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

func TwitchLoginHandler(w http.ResponseWriter, r *http.Request) {
	token, err := controllerhelpers.GetToken(r)
	if err == http.ErrNoCookie {
		http.Error(w, "You are not logged in.", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Invalid jwt", http.StatusBadRequest)
		return
	}

	steamID := token.Claims["steam_id"].(string)

	player, _ := player.GetPlayerBySteamID(steamID)

	loginURL := url.URL{
		Scheme: "https",
		Host:   "api.twitch.tv",
		Path:   "kraken/oauth2/authorize",
	}

	//twitchRedirectURL := config.Constants.PublicAddress + "/" + "twitchAuth"
	twitchRedirectURL, _ := url.Parse(config.Constants.PublicAddress)
	twitchRedirectURL.Path = "twitchAuth"

	values := loginURL.Query()
	values.Set("response_type", "code")
	values.Set("client_id", config.Constants.TwitchClientID)
	values.Set("redirect_uri", twitchRedirectURL.String())
	values.Set("scope", "channel_check_subscription user_subscriptions channel_subscriptions user_read")
	values.Set("state", xsrftoken.Generate(config.Constants.CookieStoreSecret, player.SteamID, "GET"))
	loginURL.RawQuery = values.Encode()

	http.Redirect(w, r, loginURL.String(), http.StatusTemporaryRedirect)
}

func TwitchAuthHandler(w http.ResponseWriter, r *http.Request) {
	token, err := controllerhelpers.GetToken(r)
	if err == http.ErrNoCookie {
		http.Error(w, "You are not logged in.", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Invalid jwt", http.StatusBadRequest)
		return
	}

	steamID := token.Claims["steam_id"].(string)
	player, _ := player.GetPlayerBySteamID(steamID)

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

	twitchRedirectURL, _ := url.Parse(config.Constants.PublicAddress)
	twitchRedirectURL.Path = "twitchAuth"

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
	values.Set("redirect_uri", twitchRedirectURL.String())
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

func TwitchLogoutHandler(w http.ResponseWriter, r *http.Request) {
	token, err := controllerhelpers.GetToken(r)
	if err == http.ErrNoCookie {
		http.Error(w, "You are not logged in.", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Invalid jwt", http.StatusBadRequest)
		return
	}

	steamID := token.Claims["steam_id"].(string)

	player, _ := player.GetPlayerBySteamID(steamID)
	player.TwitchName = ""
	player.TwitchAccessToken = ""
	player.Save()

	referer, ok := r.Header["Referer"]
	if ok {
		http.Redirect(w, r, referer[0], 303)
		return
	}

	http.Redirect(w, r, config.Constants.LoginRedirectPath, http.StatusTemporaryRedirect)
}
