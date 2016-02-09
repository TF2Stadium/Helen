package helpers

import (
	"errors"
	"net"
	"strings"

	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

var kapi client.KeysAPI

func ConnectEtcd(addr string) error {
	var err error

	cfg := client.Config{Endpoints: []string{addr}}
	c, err := client.New(cfg)
	if err != nil {
		return err
	}

	kapi = client.NewKeysAPI(c)

	return nil
}

func GetAddr(serviceName string) (string, error) {
	resp, err := kapi.Get(context.Background(), serviceName, nil)
	if err != nil {
		return "", err
	}

	return resp.Node.Value, nil
}

func SetAddr(serviceName string, addr string) (*client.Node, error) {
	l, _ := net.InterfaceAddrs()
	var ipaddr string
	for _, addr := range l {
		if ip, ok := addr.(*net.IPNet); ok && !ip.IP.IsLoopback() && ip.IP.To4() != nil {
			ipaddr = strings.Split(ip.String(), "/")[0]
			break
		}
	}

	if ipaddr == "" {
		return nil, errors.New("Couldn't get IP Address.")
	}

	if arr := strings.Split(addr, ":"); len(arr) != 0 {
		ipaddr += ":" + arr[1]
	}

	resp, err := kapi.Set(context.Background(), serviceName, ipaddr, nil)
	if err != nil {
		return nil, err
	}

	return resp.Node, nil
}
