package step_test

import (
	"strings"

	"github.com/sirupsen/logrus"
)

type testStep struct {
	err     error
	name    string
	counter *int
	logf    *logrus.Entry
}

func (e testStep) Name() string {
	return e.name
}

func (e testStep) Run() error {
	e.logf = e.logf.WithField("Step", e.name)
	return e.err
}

func (e testStep) Cleanup() error {
	return nil
}

func (e testStep) OnError() error {
	*e.counter++
	e.logf.Infof("Called on Error, resource: %s", e.name)
	return nil
}

func getLogsContains(entries []*logrus.Entry, text string) []*logrus.Entry {
	filteredEntries := []*logrus.Entry{}

	for _, entry := range entries {
		if strings.Contains(entry.Message, text) {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	return filteredEntries
}

func getLogs(entries []*logrus.Entry, key, value string) []*logrus.Entry {
	var foundLogs []*logrus.Entry
	for _, entry := range entries {
		field, ok := entry.Data[key]
		if ok {
			logValue, ok := field.(string)
			if ok {
				if logValue == value {
					foundLogs = append(foundLogs, entry)
				}
			}
		}
	}
	return foundLogs
}

func getFirstMatchingLog(entries []*logrus.Entry, text string, startIdx int) (*logrus.Entry, int) {
	for i := startIdx; i < len(entries); i++ {
		if strings.Contains(entries[i].Message, text) {
			return entries[i], i
		}
	}
	return nil, -1
}

func getLogsWithLevel(entries []*logrus.Entry, level logrus.Level) []*logrus.Entry {
	filtered := []*logrus.Entry{}
	for _, entry := range entries {
		if entry.Level == level {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}
