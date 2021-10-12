package logger

//noinspection SpellCheckingInspection
import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

const (
	timestampFormat = "2006-01-02T15:04:05.999Z"
	fieldBuildHash  = "buildHash"

	// FieldCtx is a key of logged context
	FieldCtx = "ctx"
)

// reservedFields defines all fields keys which cannot be used when
// we logging additional fields with YaaS formatter
var reservedFields = map[string]struct{}{
	"hop":         {},
	"requestId":   {},
	"time":        {},
	"message":     {},
	"requestOrg":  {},
	"requestUser": {},
}

// Formatter is a log formatter
type Formatter struct{}

// Format formats log entry to adhere to YaaS specs
func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(logrus.Fields, len(entry.Data)+3)

	for k, v := range entry.Data {
		if shouldSkipThisField(k) {
			continue
		}

		switch v := v.(type) {
		case error:
			data[k] = v.Error()
		default:
			data[k] = v
		}
	}

	data["time"] = entry.Time.Format(timestampFormat)
	data["message"] = entry.Message

	logEntry := map[string]interface{}{
		"log":   data,
		"level": entry.Level.String(),
	}

	serialized, err := json.Marshal(logEntry)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	}
	return append(serialized, '\n'), nil
}

func shouldSkipThisField(name string) bool {
	// skips because is should be process by another function
	if name == FieldCtx {
		return true
	}

	// skip because this fields are reserved
	_, found := reservedFields[name]
	return found
}

// New creates new instance of logging apparatus.
// Logrus.Entry is returned as it's always decorated with fields.
func New(cfg *Config) *logrus.Entry {

	lgr := &logrus.Logger{
		Out:       os.Stdout,
		Formatter: new(Formatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.Level(cfg.Level),
	}

	fields := logrus.Fields{}

	if cfg.BuildHash != "" {
		fields[fieldBuildHash] = cfg.BuildHash
	}

	return lgr.WithFields(fields)
}
