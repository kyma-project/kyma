package origin

import "strings"

func Match(s, pattern string) bool {
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
