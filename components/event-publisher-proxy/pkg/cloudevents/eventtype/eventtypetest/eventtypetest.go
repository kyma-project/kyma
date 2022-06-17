// Package eventtypetest provides utilities for eventype testing.
package eventtypetest

type CleanerFunc func(string) (string, error)

func (cf CleanerFunc) Clean(eventType string) (string, error) {
	return cf(eventType)
}

var DefaultCleaner = func(eventType string) (string, error) {
	return eventType, nil
}
