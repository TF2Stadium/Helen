package helpers

import (
	"net"
	"strings"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/servemetf"
)

var (
	servemeNA = &servemetf.Context{Host: "na.serveme.tf"}
	servemeEU = &servemetf.Context{Host: "serveme.tf"}
	servemeAU = &servemetf.Context{Host: "au.serveme.tf"}
)

func GetServemeContextIP(ipaddr string) *servemetf.Context {
	continent, _ := GetRegion(ipaddr)

	switch continent {
	case "NA":
		return servemeNA
	case "SA":
		return servemeNA

	case "AS":
		return servemeNA
	case "EU":
		return servemeEU

	case "OC":
		return servemeAU
	default:
		return servemeEU
	}
}

func GetServemeContext(addrStr string) *servemetf.Context {
	arr := strings.Split(addrStr, ":")
	addr, err := net.ResolveIPAddr("ip4", arr[0])
	if err != nil {
		return servemeEU
	}

	return GetServemeContextIP(addr.String())
}

func SetServemeContext() {
	servemeNA.APIKey = config.Constants.ServemeAPIKey
	servemeEU.APIKey = config.Constants.ServemeAPIKey
	servemeAU.APIKey = config.Constants.ServemeAPIKey
}
