// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package controllers

import (
	"fmt"
	"html/template"
	"net/http"
	"runtime"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/internal/version"
	"github.com/TF2Stadium/Helen/models/player"
)

var (
	mainTempl *template.Template
)

func MainHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	var p *player.Player
	token, err := controllerhelpers.GetToken(r)

	if err == nil {
		p = controllerhelpers.GetPlayer(token)
	}

	errtempl := mainTempl.Execute(w, map[string]interface{}{
		"LoggedIn":  err == nil,
		"Player":    p,
		"MockLogin": config.Constants.MockupAuth,
		"BuildDate": version.BuildDate,
		"GitCommit": version.GitCommit,
		"GitBranch": version.GitBranch,
		"BuildInfo": fmt.Sprintf("Built using %s on %s (%s %s)", runtime.Version(), version.Hostname, runtime.GOOS, runtime.GOARCH),
	})
	if errtempl != nil {
		logrus.Error(err)
	}
}
