package admin

import (
	"html/template"
	"net/http"

	"github.com/TF2Stadium/Helen/helpers"
)

func confirmReq(w http.ResponseWriter, url, redirect string, title string) {
	templ, err := template.ParseFiles("views/admin/templates/confirm.html")
	if err != nil {
		helpers.Logger.Error(err.Error())
		return
	}

	templ.Execute(w, struct {
		URL         string
		RedirectURL string
		Title       string
	}{url, redirect, title})
}
