// Package eventtypetest provides utilities for eventype testing.
package eventtypetest

type CleanerFunc func(string) (string, error)

type CleanerStub struct {
	CleanType string
	Error     error
}

func (c CleanerStub) Clean(_ string) (string, error) {
	return c.CleanType, c.Error
}

func (cf CleanerFunc) Clean(eventType string) (string, error) {
	return cf(eventType)
}

var DefaultCleaner = func(eventType string) (string, error) {
	return eventType, nil
}
