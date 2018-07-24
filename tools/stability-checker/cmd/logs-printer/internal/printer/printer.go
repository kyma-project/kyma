package printer

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/kyma-project/kyma/tools/stability-checker/internal/runner"
)

const (
	invertColor = "\033[7m"
	noColor     = "\033[0m"
)

// LogPrinter prints stability-checker logs
type LogPrinter struct {
	stream           *json.Decoder
	requestedTestIDs map[string]struct{}
}

// New returns new instance of LogPrinter
func New(stream *json.Decoder, ids []string) *LogPrinter {
	var mapped map[string]struct{}
	if len(ids) != 0 {
		mapped = make(map[string]struct{})
		for _, id := range ids {
			mapped[id] = struct{}{}
		}
	}

	return &LogPrinter{
		stream:           stream,
		requestedTestIDs: mapped,
	}
}

// PrintFailedTestOutput prints failed tests outputs.
func (l *LogPrinter) PrintFailedTestOutput() error {
	for {
		var e logEntry
		switch err := l.stream.Decode(&e); err {
		case nil:
		case io.EOF:
			return nil
		default:
			return err
		}

		if l.shouldSkipLogMsg(e) {
			continue
		}

		fmt.Print(invertColor)
		fmt.Printf("[%s] Output for test id %q", e.Log.Time, e.Log.TestRunID)
		fmt.Print(noColor)

		fmt.Printf("\n %s \n", e.Log.Message)
	}
}

func (l *LogPrinter) shouldSkipLogMsg(entry logEntry) bool {
	if entry.Level != "error" {
		return true
	}

	if entry.Log.Type != runner.TestOutputLogType {
		return true
	}

	if l.requestedTestIDs != nil {
		if _, found := l.requestedTestIDs[entry.Log.TestRunID]; !found {
			return true
		}
	}

	return false
}
