// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

import (
	"io"
	"net"
	"net/rpc"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/TF2Stadium/Helen/config"
)

func isNetworkError(err error) bool {
	_, ok := err.(*net.OpError)
	return ok || err == io.ErrUnexpectedEOF || err == rpc.ErrShutdown

}

func connect(addr string) *rpc.Client {
	client, err := rpc.DialHTTP("tcp", addr)
	for err != nil {
		logrus.Error(err.Error())
		time.Sleep(1 * time.Second)
		client, err = rpc.DialHTTP("tcp", addr)
	}

	return client
}

var (
	rpcClientMap = make(map[string]*rpc.Client)
	mu           = new(sync.RWMutex)
)

func ConnectRPC() {
	if config.Constants.PaulingAddr != "" {
		client := connect(config.Constants.PaulingAddr)
		rpcClientMap[config.Constants.PaulingAddr] = client
		logrus.Info("Connected to Pauling on port ", config.Constants.PaulingAddr)
	}
	if config.Constants.FumbleAddr != "" {
		client := connect(config.Constants.FumbleAddr)
		rpcClientMap[config.Constants.FumbleAddr] = client
		logrus.Info("Connected to Fumble on port ", config.Constants.FumbleAddr)
	}
}

func call(addr, method string, args, reply interface{}) error {
	if addr == "" {
		return nil
	}

	mu.RLock()
	client := rpcClientMap[addr]
	mu.RUnlock()

	err := client.Call(method, args, reply)
	if isNetworkError(err) {
		mu.Lock()
		rpcClientMap[addr] = connect(addr)
		mu.Unlock()
	}

	return err
}
