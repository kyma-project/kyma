package utils

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// GetPortNumberFromURL converts string port from url.URL to uint32 port.
func GetPortNumberFromURL(u url.URL) (uint32, error) {
	port := uint32(0)
	sinkPort := u.Port()
	if sinkPort != "" {
		u64, err := strconv.ParseUint(sinkPort, 10, 32)
		if err != nil {
			return port, errors.Wrapf(err, "failed to convert port: %s", u.Port())
		}
		port = uint32(u64)
	}
	if port == uint32(0) {
		switch strings.ToLower(u.Scheme) {
		case "https":
			port = uint32(443)
		default:
			port = uint32(80)
		}
	}
	return port, nil
}
