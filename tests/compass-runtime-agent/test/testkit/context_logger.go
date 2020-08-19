package testkit

import (
	"fmt"
	"testing"
)

type Logger struct {
	t            *testing.T
	fields       map[string]string
	joinedFields string
}

func NewLogger(t *testing.T, fields map[string]string) *Logger {
	return &Logger{
		t:            t,
		fields:       fields,
		joinedFields: joinFields(fields),
	}
}

func (l Logger) NewExtended(fields map[string]string) *Logger {
	mergedFields := mergeFields(l.fields, fields)

	return &Logger{
		t:            l.t,
		fields:       mergedFields,
		joinedFields: joinFields(mergedFields),
	}
}

func (l Logger) Log(msg string) {
	l.t.Log(l.ContextMsg(msg))
}

func (l Logger) ContextMsg(msg string) string {
	return fmt.Sprintf("%s   %s", msg, l.joinedFields)
}

func (l *Logger) AddField(key, val string) {
	l.fields[key] = val
	l.joinedFields = joinFields(l.fields)
}

func joinFields(fields map[string]string) string {
	str := ""

	for k, v := range fields {
		str = fmt.Sprintf("%s | %s=%s", str, k, v)
	}
	return str
}

func mergeFields(old, new map[string]string) map[string]string {
	merged := map[string]string{}

	for k, v := range old {
		merged[k] = v
	}

	for k, v := range new {
		merged[k] = v
	}

	return merged
}
