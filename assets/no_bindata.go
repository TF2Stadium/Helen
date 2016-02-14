// +build !bindata

package assets

import (
	"io/ioutil"

	"github.com/Sirupsen/logrus"
)

func MustAsset(name string) []byte {
	bytes, err := ioutil.ReadFile(name)
	if err != nil {
		logrus.Fatal(err)
	}

	return bytes
}
