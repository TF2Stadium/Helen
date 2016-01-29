package admin

import (
	"html/template"
	"net/http"

	"github.com/Sirupsen/logrus"
)

func confirmReq(w http.ResponseWriter, url, redirect string, title string) {
	templ, err := template.ParseFiles("views/admin/templates/confirm.html")
	if err != nil {
		logrus.Error(err.Error())
		return
	}

	templ.Execute(w, struct {
		URL         string
		RedirectURL string
		Title       string
	}{url, redirect, title})
}
