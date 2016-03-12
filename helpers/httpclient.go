package helpers

import (
	"net/http"
	"time"
)

var HTTPClient = &http.Client{
	Timeout: 20 * time.Second,
}
