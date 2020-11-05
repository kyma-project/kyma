package utils

import (
	"github.com/pkg/errors"
	"net/url"
	"strconv"
	"strings"
)

// ConvertURLPortForApiRulePort converts string port from url.URL to uint32 port
func ConvertStringPortUInt32Port(u url.URL) (uint32, error) {
	port := uint32(0)
	sinkPort := u.Port()
	if sinkPort != "" {
		u64, err := strconv.ParseUint(u.Port(), 10, 32)
		if err != nil {
			return port, errors.Wrapf(err, "failed to convert port: %s", u.Port())
		}
		port = uint32(u64)
	}
	if port == uint32(0) {
		switch strings.ToLower(u.Scheme) {
		case "http":
			port = uint32(80)
		case "https":
			port = uint32(443)
		}
	}
	return port, nil
}
