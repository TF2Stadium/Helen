package helpers

import (
	"net"
	"strings"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/servemetf"
)

var (
	ServemeNA = &servemetf.Context{Host: "na.serveme.tf"}
	ServemeEU = &servemetf.Context{Host: "serveme.tf"}
	ServemeAU = &servemetf.Context{Host: "au.serveme.tf"}
)

func GetServemeContextIP(ipaddr string) *servemetf.Context {
	continent, _ := GetRegion(ipaddr)
	switch strings.ToLower(continent) {
	case "na": // north america
		return ServemeNA
	case "sa": // south america
		return ServemeNA

	case "as": // asia
		return ServemeEU
	case "eu": // europe
		return ServemeEU

	case "oc": // oceania
		return ServemeAU

	default: // africa and antarctica
		return ServemeEU
	}
}

func GetServemeContext(addrStr string) *servemetf.Context {
	arr := strings.Split(addrStr, ":")
	addr, err := net.ResolveIPAddr("ip4", arr[0])
	if err != nil {
		return ServemeEU
	}

	return GetServemeContextIP(addr.String())
}

func init() {
	ServemeNA.APIKey = config.Constants.ServemeAPIKey
	ServemeEU.APIKey = config.Constants.ServemeAPIKey
	ServemeAU.APIKey = config.Constants.ServemeAPIKey
}
