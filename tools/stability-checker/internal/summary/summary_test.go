package summary_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/kyma-project/kyma/tools/stability-checker/internal/log"
	"github.com/kyma-project/kyma/tools/stability-checker/internal/summary"
	"github.com/kyma-project/kyma/tools/stability-checker/internal/summary/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestForGettingTestSummaryForExecutions(t *testing.T) {
	// GIVEN
	mockLogFetcher := &automock.LogFetcher{}
	defer mockLogFetcher.AssertExpectations(t)
	mockLogProcessor := &automock.LogProcessor{}
	defer mockLogProcessor.AssertExpectations(t)

	mockLogProcessor.On("Process", []byte("FAILED test-a. SUCCESS test-b.")).Return(map[string]summary.SpecificTestStats{
		"test-a": {
			Name:     "test-a",
			Failures: 1,
		},
		"test-b": {
			Name:      "test-b",
			Successes: 1,
		}}, nil).Times(3)
	mockLogFetcher.On("GetLogsFromPod").Return(fixReadCloser(), nil).Once()

	sut := summary.NewService(mockLogFetcher, mockLogProcessor)
	// WHEN
	stats, err := sut.GetTestSummaryForExecutions(fixTestIDs())
	// THEN
	require.NoError(t, err)
	assert.Equal(t, fixResults(), stats)
}

func fixTestIDs() []string {
	return []string{
		"id-0",
		"id-2",
		"id-3",
	}
}

func fixResults() []summary.SpecificTestStats {
	return []summary.SpecificTestStats{
		{
			Name:     "test-a",
			Failures: 3,
		},
		{
			Name:      "test-b",
			Successes: 3,
		},
	}
}

func fixLogEntries() []log.Entry {
	out := make([]log.Entry, 4)
	for i := 0; i < 4; i++ {
		e := log.Entry{}
		e.Log.Message = "FAILED test-a. SUCCESS test-b."
		e.Log.TestRunID = fmt.Sprintf("id-%d", i)
		out[i] = e
	}

	return out
}

func fixReadCloser() io.ReadCloser {
	buff := bytes.NewBuffer(nil)
	for _, e := range fixLogEntries() {
		b, _ := json.Marshal(e)
		buff.Write(b)
	}

	out := &fakeReadCloser{
		Buffer: buff,
	}
	return out
}

type fakeReadCloser struct {
	*bytes.Buffer
	CloseCalled bool
}

func (m *fakeReadCloser) Close() error {
	m.CloseCalled = true
	return nil
}
