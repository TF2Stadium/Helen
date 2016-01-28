package pprof

import (
	"expvar"
	"runtime/pprof"
)

//Clients stores the number of connected clients
var Clients = new(expvar.Int)

func goroutines() interface{} {
	n := pprof.Lookup("goroutine").Count()
	return n
}

func init() {
	expvar.Publish("goroutines", expvar.Func(goroutines))
	expvar.Publish("clients", Clients)
}
