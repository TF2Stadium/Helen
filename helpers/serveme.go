package helpers

import (
	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/servemetf"
)

var ServemeContext *servemetf.Context

func SetServemeContext() {
	ServemeContext = new(servemetf.Context)
	ServemeContext.APIKey = config.Constants.ServemeAPIKey
}
