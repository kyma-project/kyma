package summary

import (
	"encoding/json"
	"io"

	"github.com/kyma-project/kyma/tools/stability-checker/internal/log"
	"github.com/pkg/errors"
)

// Service is responsible for producing summary for test executions.
type Service struct {
	logFetcher logFetcher
	processor  logProcessor
}

// NewService returns Service
func NewService(logFetcher logFetcher, processor logProcessor) *Service {
	return &Service{
		logFetcher: logFetcher,
		processor:  processor,
	}
}

// GetTestSummaryForExecutions analyzes logs from test executions and produces summary for specific tests.
func (c *Service) GetTestSummaryForExecutions(testIDs []string) ([]SpecificTestStats, error) {
	readCloser, err := c.logFetcher.GetLogsFromPod()
	if err != nil {
		return nil, err
	}
	defer readCloser.Close()
	stream := json.NewDecoder(readCloser)

	testIDMap := func(ids []string) map[string]struct{} {
		out := map[string]struct{}{}
		for _, id := range ids {
			out[id] = struct{}{}
		}
		return out
	}(testIDs)

loop:
	for {
		var e log.Entry
		switch err := stream.Decode(&e); err {
		case nil:
		case io.EOF:
			break loop
		default:
			return nil, errors.Wrap(err, "while decoding stream")
		}

		_, contains := testIDMap[e.Log.TestRunID]
		if contains {
			if err := c.processor.Process([]byte(e.Log.Message)); err != nil {
				return nil, errors.Wrap(err, "while processing test output")
			}
		}

	}

	return c.processor.GetResults(), nil
}

// dependencies

//go:generate mockery -name=logProcessor -output=automock -outpkg=automock -case=underscore
type logProcessor interface {
	Process([]byte) error
	GetResults() []SpecificTestStats
}

//go:generate mockery -name=logFetcher -output=automock -outpkg=automock -case=underscore
type logFetcher interface {
	GetLogsFromPod() (io.ReadCloser, error)
}