package helpers

import (
	"strings"

	"github.com/Sirupsen/logrus"
)

type Address struct {
	Address string
}

func (d Address) String() string {
	addr := d.Address
	var err error

	if strings.HasPrefix(addr, "etcd:") {
		addr, err = GetAddr(strings.Split(addr, ":")[1])
		if err != nil {
			logrus.Fatal(err)
		}
	}

	return addr
}
