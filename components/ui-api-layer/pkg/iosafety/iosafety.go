package iosafety

import (
	"io"
	"io/ioutil"
)

// DrainReader reads and discards the remaining part in reader (for example response body data)
// In case of HTTP this ensured that the http connection could be reused for another request if the keepalive http connection behavior is enabled.
func DrainReader(reader io.Reader) error {
	if reader == nil {
		return nil
	}
	_, drainError := io.Copy(ioutil.Discard, io.LimitReader(reader, 4096))
	return drainError
}
