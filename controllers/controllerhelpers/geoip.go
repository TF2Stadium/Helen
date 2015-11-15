package controllerhelpers

import (
	"net"
	"path/filepath"
	"strings"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/oschwald/geoip2-golang"
)

var geodb *geoip2.Reader

func InitDB() {
	if config.Constants.GeoIP == "" {
		return
	}

	var err error
	filename, _ := filepath.Abs(config.Constants.GeoIP)
	geodb, err = geoip2.Open(filename)

	if err != nil {
		helpers.Logger.Fatal(err.Error())
	}
}

func GetRegion(server string) string {
	if config.Constants.GeoIP == "" {
		return ""
	}

	arr := strings.Split(server, ":")
	addr, err := net.ResolveIPAddr("ip4", arr[0])
	if err != nil {
		helpers.Logger.Error(err.Error())
		return ""

	}

	record, err := geodb.Country(addr.IP)
	if err != nil {
		helpers.Logger.Error(err.Error())
		return ""
	}
	return record.Continent.Code
}
