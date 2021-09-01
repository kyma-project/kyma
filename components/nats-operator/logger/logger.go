package logger

import "github.com/sirupsen/logrus"

const (
	// LogKeyReason is used as a named key for a log message with reason.
	LogKeyReason = "reason"

	// LogKeySolution is used as a named key for a log message with solution.
	LogKeySolution = "solution"
)

// New returns a new Logger instance.
func New() *logrus.Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	return log
}
