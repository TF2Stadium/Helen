package admin

import (
	"html/template"
)

func InitAdminTemplates() {
	serverPage = template.Must(template.ParseFiles("views/admin/templates/server.html"))
	banlogsTempl = template.Must(template.ParseFiles("views/admin/templates/ban_logs.html"))
	chatLogsTempl = template.Must(template.ParseFiles("views/admin/templates/chatlogs.html"))
	lobbiesTempl = template.Must(template.ParseFiles("views/admin/templates/lobbies.html"))
	adminPageTempl = template.Must(template.ParseFiles("views/admin/index.html"))
}
