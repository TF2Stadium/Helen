package admin

import (
	"errors"
	"html/template"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/models"
	"golang.org/x/net/xsrftoken"
)

func confirmReq(w http.ResponseWriter, r *http.Request, method, title string) {
	templ, err := template.ParseFiles("views/admin/templates/confirm.html")
	if err != nil {
		logrus.Error(err.Error())
		return
	}

	session, _ := controllerhelpers.GetSessionHTTP(r)
	admin, _ := models.GetPlayerBySteamID(session.Values["steam_id"].(string))
	token := xsrftoken.Generate(config.Constants.CookieStoreSecret, admin.SteamID, method)

	templ.Execute(w, struct {
		URL       string
		Title     string
		XSRFToken string
	}{r.URL.String(), title, token})
}

func verifyToken(r *http.Request, method string) error {
	token := r.Form.Get("xsrf-token")
	if token == "" {
		return errors.New("No XSRF token present in form")
	}

	session, _ := controllerhelpers.GetSessionHTTP(r)
	admin, _ := models.GetPlayerBySteamID(session.Values["steam_id"].(string))

	valid := xsrftoken.Valid(token, config.Constants.CookieStoreSecret, admin.SteamID, method)

	if !valid {
		return errors.New("XSRF token is invalid")
	}

	return nil
}
