package config

import (
	"path/filepath"

	"github.com/N1xx1/golang-i18n"
	"github.com/TF2Stadium/Helen/helpers"
)

func InitializeLocalizations(fileName string) {
	realPath, err := filepath.Abs(fileName)
	if err != nil {
		helpers.Logger.Fatal(err.Error())
	}
	err = i18n.GlobalTfunc(realPath)
	if err != nil {
		helpers.Logger.Fatal(err.Error())
	}
}
