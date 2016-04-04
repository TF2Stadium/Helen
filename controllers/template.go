package controllers

import (
	"html/template"

	"github.com/TF2Stadium/Helen/controllers/admin"
)

func InitTemplates() {
	admin.InitAdminTemplates()

	twitchBadge = template.Must(template.ParseFiles("views/twitchbadge.html"))
	mainTempl = template.Must(template.ParseFiles("views/index.html"))
}
