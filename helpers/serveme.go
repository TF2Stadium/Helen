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
	switch strings.ToLower(continent) {
	case "na": // north america
		return servemeNA
	case "sa": // south america
		return servemeNA

	case "as": // asia
		return servemeEU
	case "eu": // europe
		return servemeEU

	case "oc": // oceania
		return servemeAU

	default: // africa and antarctica
		return servemeNA
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
