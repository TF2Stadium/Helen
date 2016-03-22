package stats

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

var Stats struct {
	NAPlayers, EUPlayers, AUPlayers, ASPlayers *int64
	Clients                                    *int64
}

func init() {
	Stats.NAPlayers = new(int64)
	Stats.EUPlayers = new(int64)
	Stats.AUPlayers = new(int64)
	Stats.ASPlayers = new(int64)
	Stats.Clients = new(int64)
}

func StatsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `{
"na": %d,
"eu": %d,
"au": %d,
"as": %d,
"clients": %d,
}`,
		atomic.LoadInt64(Stats.NAPlayers),
		atomic.LoadInt64(Stats.EUPlayers),
		atomic.LoadInt64(Stats.AUPlayers),
		atomic.LoadInt64(Stats.ASPlayers),
		atomic.LoadInt64(Stats.Clients))
}