package controllerhelpers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/models/player"
	"github.com/dgrijalva/jwt-go"
)

var (
	signingKey []byte
)

func SetupJWTSigning() {
	if config.Constants.CookieStoreSecret == "secret" {
		logrus.Warning("Using an insecure encryption key")
		signingKey = []byte("secret")
		return
	}

	var err error
	signingKey, err = base64.StdEncoding.DecodeString(config.Constants.CookieStoreSecret)
	if err != nil {
		logrus.Fatal(err)
	}
}

func NewToken(player *player.Player) string {
	token := jwt.New(jwt.SigningMethodHS512)
	token.Claims["player_id"] = strconv.FormatUint(uint64(player.ID), 10)
	token.Claims["steam_id"] = player.SteamID
	token.Claims["mumble_password"] = player.MumbleAuthkey
	token.Claims["role"] = strconv.Itoa(int(player.Role))
	token.Claims["iat"] = time.Now().Unix()
	token.Claims["iss"] = config.Constants.PublicAddress

	str, err := token.SignedString([]byte(signingKey))
	if err != nil {
		logrus.Error(err)
	}

	return str
}

func verifyToken(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
	}

	return signingKey, nil
}

func GetToken(r *http.Request) (*jwt.Token, error) {
	cookie, err := r.Cookie("auth-jwt")
	if err != nil {
		return nil, err
	}

	token, err := jwt.Parse(cookie.Value, verifyToken)
	return token, err
}

func GetPlayer(token *jwt.Token) *player.Player {
	playerid, _ := strconv.ParseUint(token.Claims["player_id"].(string), 10, 32)
	player, _ := player.GetPlayerByID(uint(playerid))
	return player
}
