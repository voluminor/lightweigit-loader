package lightweigit

import (
	"errors"
	"net/http"
	"time"
)

// // // // // // // // // // // // // // // //

var (
	HttpClient = &http.Client{Timeout: 4 * time.Second}

	ErrNotFound = errors.New("not found")
	ErrModTag   = errors.New("invalid tag")
)
