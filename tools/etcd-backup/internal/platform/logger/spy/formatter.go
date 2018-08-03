package spy

import (
	"sync"

	"github.com/sirupsen/logrus"
)

// EntryAssertFormatter is a log formatter, which gather all logged entries
type EntryAssertFormatter struct {
	entries    []logrus.Entry
	Underlying logrus.Formatter
	mu         sync.RWMutex
}

// Format appends each entry to entries slice
func (f *EntryAssertFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	f.appendThreadSafe(*entry)
	return f.Underlying.Format(entry)
}

// AnyMatches iterates over all stored entries and execute given matcher on it.
// Return true if any entries was successful matched
func (f *EntryAssertFormatter) AnyMatches(matcher func(entry logrus.Entry) bool) bool {
	f.mu.RLock()
	copyOfEntries := make([]logrus.Entry, len(f.entries))
	copy(copyOfEntries, f.entries)
	f.mu.RUnlock()
	for _, entry := range copyOfEntries {
		if matcher(entry) {
			return true
		}
	}
	return false
}

// AllEntriesMatches iterates over all stored entries and execute given matcher on it.
// Return true only if all entries was successful matched
func (f *EntryAssertFormatter) AllEntriesMatches(matcher func(entry logrus.Entry) bool) bool {
	f.mu.RLock()
	copyOfEntries := make([]logrus.Entry, len(f.entries))
	copy(copyOfEntries, f.entries)
	f.mu.RUnlock()
	for _, entry := range copyOfEntries {
		if !matcher(entry) {
			return false
		}
	}
	return true
}

func (f *EntryAssertFormatter) appendThreadSafe(entry logrus.Entry) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.entries = append(f.entries, entry)
}
