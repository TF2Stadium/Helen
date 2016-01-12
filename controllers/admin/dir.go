// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package admin

import (
	"net/http"
	"path/filepath"
)

func ServeAdminPage(w http.ResponseWriter, r *http.Request) {
	abs, _ := filepath.Abs("./views/admin/ban")
	http.ServeFile(w, r, abs)
}
