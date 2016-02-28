package helpers

import (
	"sync"
)

//GlobalWait is used by the main signal handler to wait for all
//state changes to finish, so as to avoid inconsistent state
//changes
var GlobalWait = new(sync.WaitGroup)
