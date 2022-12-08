package builder

import (
	"fmt"
	"regexp"
	"strings"
)

// removeNonAlphanumeric takes a string,removes all non-alphanumerical character besides dot (".") and returns the clean
// version.
func removeNonAlphanumeric(s string) string {
	return regexp.MustCompile("[^a-zA-Z0-9.]+").ReplaceAllString(s, "")
}

func consolidateToMaxNumberOfSegments(maxSegments int, s string) string {
	segments := strings.Split(s, ".")

	if len(segments) <= maxSegments {
		return concatSegmentsWithDot(segments...)
	}

	cut := len(segments) - maxSegments
	first := concatSegmentsWithString("", segments[:cut]...)
	last := concatSegmentsWithDot(segments[cut:]...)

	return fmt.Sprintf("%s%s", first, last)
}

func concatSegmentsWithDot(segments ...string) string {
	return concatSegmentsWithString(".", segments...)
}

// concatSegmentsWithCharacter takes an array of strings and concatenate them with a string in
// between. For example "." and ["a", "b", "c", "d", "e", "f"] would lead to
// "a.b.c.d.e.f".
func concatSegmentsWithString(character string, segments ...string) string {
	s := ""
	for _, segment := range segments {
		if s == "" {
			s = segment
			continue
		}
		if segment != "" {
			s = fmt.Sprintf("%s%s%s", s, character, segment)
		}
	}
	return s
}
