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

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/helpers"
)

func isNetworkError(err error) bool {
	_, ok := err.(*net.OpError)
	return ok || err == io.ErrUnexpectedEOF || err == rpc.ErrShutdown

}

func connect(port string) *rpc.Client {
	client, err := rpc.DialHTTP("tcp", "localhost:"+port)
	for err != nil {
		helpers.Logger.Error(err.Error())
		time.Sleep(1 * time.Second)
		client, err = rpc.DialHTTP("tcp", "localhost:"+port)
	}

	return client
}

var (
	rpcClientMap = make(map[string]*rpc.Client)
	mu           = new(sync.RWMutex)
)

func ConnectRPC() {
	if !config.Constants.ServerMockUp {
		client := connect(config.Constants.PaulingPort)
		rpcClientMap[config.Constants.PaulingPort] = client
	}
	if config.Constants.FumblePort != "" {
		client := connect(config.Constants.FumblePort)
		rpcClientMap[config.Constants.FumblePort] = client
	}
}

func call(port, method string, args, reply interface{}) error {
	mu.RLock()
	client := rpcClientMap[port]
	mu.RUnlock()

	err := client.Call(method, args, reply)
	if isNetworkError(err) {
		mu.Lock()
		rpcClientMap[port] = connect(port)
		mu.Unlock()
	}

	return err
}
