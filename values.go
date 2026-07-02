package lightweigit

import (
	"errors"
	"net/http"
	"time"
)

// // // // // // // // // // // // // // // //

// maxJSONBody caps a successful GetJSON body; anything larger is reported
// as ErrResponseTooLarge instead of being truncated and mis-decoded.
const maxJSONBody = 8 << 20

var (
	HttpClient = &http.Client{Timeout: 4 * time.Second}

	ErrNotFound         = errors.New("not found")
	ErrModTag           = errors.New("invalid tag")
	ErrResponseTooLarge = errors.New("response too large")
)
