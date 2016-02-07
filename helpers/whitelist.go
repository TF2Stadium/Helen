package helpers

import (
	"encoding/xml"
	"io/ioutil"

	"github.com/Sirupsen/logrus"
)

type GroupXML struct {
	//XMLName xml.Name `xml:"memberList"`
	//GroupID uint64   `xml:"groupID64"`
	Members []string `xml:"members>steamID64"`
}

// Get Steam IDs of the members of a given steam group
func GetGroupMembers(url string) ([]string, error) {
	var groupXML GroupXML

	resp, err := HTTPClient.Get(url)
	if err != nil {
		logrus.Error(err.Error())
		return []string{}, err
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Error(err.Error())
		return []string{}, err
	}

	//	Logger.Debug(string(bytes))
	xml.Unmarshal(bytes, &groupXML)
	//Logger.Debug("%v", groupXML.Members)

	return groupXML.Members, nil
}

func IsWhitelisted(steamid, url string) bool {
	whitelist, _ := GetGroupMembers(url)

	for _, steamID := range whitelist {
		//Logger.Debug("%s %s", steamid, steamID)
		if steamid == steamID {
			return true
		}
	}

	return false
}
