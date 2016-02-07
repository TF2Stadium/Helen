package helpers

import (
	"net/http"
	"time"
)

var HTTPClient = &http.Client{
	Timeout: 5 * time.Second,
}
