package logger

import "github.com/sirupsen/logrus"

const (
	// LogKeyReason is used to log a reason for an issue.
	LogKeyReason = "reason"

	// LogKeySolution is used to log a solution for an issue.
	LogKeySolution = "solution"
)

// New returns a new Logger instance.
func New() *logrus.Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	return log
}
