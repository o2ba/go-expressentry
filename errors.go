package expressentry

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	// ErrUnexpectedFormat is wrapped by the error returned when a file isn't
	// a rounds file, or none of its rounds can be read.
	ErrUnexpectedFormat = errors.New("expressentry: unexpected format")
)

// HTTPError is returned when a rounds file answers with a status other than
// 200 OK.
type HTTPError struct {
	URL        string
	StatusCode int
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("expressentry: GET %s: %d %s", e.URL, e.StatusCode, http.StatusText(e.StatusCode))
}
