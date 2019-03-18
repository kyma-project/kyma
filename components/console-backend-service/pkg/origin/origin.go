package origin

import (
	"net/http"
	"strings"
)

func CheckFn(allowedOrigins []string) func(r *http.Request) bool {
	return func(r *http.Request) bool {
		if r == nil {
			return false
		}

		requestOrigin := r.Header.Get("Origin")
		if requestOrigin == "" {
			return true
		}

		for _, allowed := range allowedOrigins {
			if match(requestOrigin, allowed) {
				return true
			}
		}

		return false
	}
}

func match(s, pattern string) bool {
	left, right := split(pattern)
	return strings.HasPrefix(s, left) && strings.HasSuffix(s, right)
}

func split(pattern string) (string, string) {
	spliced := strings.SplitN(pattern, "*", 2)

	if len(spliced) == 2 {
		return spliced[0], spliced[1]
	}

	if strings.HasPrefix(pattern, "*") {
		return "", spliced[0]
	}

	return spliced[0], ""
}
