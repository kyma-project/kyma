package logger

import (
	"time"

	"github.com/sirupsen/logrus"
)

// THTimeForcedFormatter is a Logrus compatible formatter which wraps original formatter and forces specific time.
// Designed to be used in testing.
type THTimeForcedFormatter struct {
	// OrigFormatter is an original formatter
	OrigFormatter logrus.Formatter

	// Time is a time to be forces in entries.
	Time time.Time
}

// Format entry but forces time for testing purposes
func (f *THTimeForcedFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(logrus.Fields, len(entry.Data))
	for k, v := range entry.Data {
		data[k] = v
	}
	entryModified := &logrus.Entry{
		Logger:  entry.Logger,
		Data:    data,
		Time:    f.Time,
		Level:   entry.Level,
		Message: entry.Message,
	}

	return f.OrigFormatter.Format(entryModified)
}
