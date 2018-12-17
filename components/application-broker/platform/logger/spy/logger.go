// Package spy provides an implementation of go-sdk.logger that helps test logging.
package spy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

// LogSink is a helper construct for testing logging in unit tests.
// Beware: all methods are working on copies of of original messages buffer and are safe for multiple uses.
type LogSink struct {
	buffer    *bytes.Buffer
	RawLogger *logrus.Logger
	Logger    *logrus.Entry
}

// NewLogSink is a factory for LogSink
func NewLogSink() *LogSink {
	buffer := bytes.NewBuffer([]byte(""))

	rawLgr := &logrus.Logger{
		Out: buffer,
		// standard json formatter is used to ease testing
		Formatter: new(logrus.JSONFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}

	lgr := rawLgr.WithField("testing", true)

	return &LogSink{
		buffer:    buffer,
		RawLogger: rawLgr,
		Logger:    lgr,
	}
}

// AssertErrorLogged checks whatever a specific string was logged as error.
//
// Compared elements: level, message
//
// Wrapped errors are supported as long as original error message ends up in resulting one.
func (s *LogSink) AssertErrorLogged(t *testing.T, errorExpected error) {
	if !s.wasLogged(t, logrus.ErrorLevel, errorExpected.Error()) {
		t.Errorf("error was not logged, expected: \"%s\"", errorExpected.Error())
	}
}

// AssertLogged checks whatever a specific string was logged at a specific level.
//
// Compared elements: level, message
//
// Beware: we are checking for sub-strings and not for the exact match.
func (s *LogSink) AssertLogged(t *testing.T, level logrus.Level, message string) {
	if !s.wasLogged(t, level, message) {
		t.Errorf("message was not logged, message: \"%s\", level: %s", message, level)
	}
}

// AssertNotLogged checks whatever a specific string was not logged at a specific level.
//
// Compared elements: level, message
//
// Beware: we are checking for sub-strings and not for the exact match.
func (s *LogSink) AssertNotLogged(t *testing.T, level logrus.Level, message string) {
	if s.wasLogged(t, level, message) {
		t.Errorf("message was logged, message: \"%s\", level: %s", message, level)
	}
}

// wasLogged checks whatever a message was logged.
//
// Compared elements: level, message
func (s *LogSink) wasLogged(t *testing.T, level logrus.Level, message string) bool {
	// new reader is created so we are safe for multiple reads
	buf := bytes.NewReader(s.buffer.Bytes())
	scanner := bufio.NewScanner(buf)
	var entryPartial struct {
		Level string `json:"level"`
		Msg   string `json:"msg"`
	}

	for scanner.Scan() {
		line := scanner.Text()

		err := json.Unmarshal([]byte(line), &entryPartial)
		if err != nil {
			t.Fatalf("unexpected error on log line unmarshalling, line: %s", line)
		}

		levelMatches := entryPartial.Level == level.String()

		// We are looking only if expected is contained (as opposed to exact match check),
		// so that e.g. errors wrapping is supported.
		containsMessage := strings.Contains(entryPartial.Msg, message)

		if levelMatches && containsMessage {
			return true
		}
	}

	return false
}

// DumpAll returns all logged messages.
func (s *LogSink) DumpAll() []string {
	// new reader is created so we are safe for multiple reads
	buf := bytes.NewReader(s.buffer.Bytes())
	scanner := bufio.NewScanner(buf)

	out := []string{}
	for scanner.Scan() {
		out = append(out, scanner.Text())
	}

	return out
}

// NewLogDummy returns dummy logger which discards logged messages on the fly.
// Useful when logger is required as dependency in unit testing.
func NewLogDummy() *logrus.Entry {
	rawLgr := logrus.New()
	rawLgr.Out = ioutil.Discard
	lgr := rawLgr.WithField("testing", true)

	return lgr
}
