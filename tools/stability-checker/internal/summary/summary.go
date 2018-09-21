package summary

import (
	"encoding/json"
	"io"

	"github.com/kyma-project/kyma/tools/stability-checker/internal/log"
	"github.com/pkg/errors"
)

// dependencies
type logProcessor interface {
	Process([]byte) error
	GetResults() []SpecificTestStats
}

type logFetcher interface {
	GetLogsFromPod() (io.ReadCloser, error)
}

// NewService ...
func NewService(logFetcher logFetcher, processor logProcessor) *Service {
	return &Service{
		logFetcher: logFetcher,
		processor:  processor,
	}
}

// Service ...
type Service struct {
	logFetcher logFetcher
	processor  logProcessor
}

// GetTestSummaryForExecutions ...
func (c *Service) GetTestSummaryForExecutions(testIDs []string) ([]SpecificTestStats, error) {
	readCloser, err := c.logFetcher.GetLogsFromPod()
	if err != nil {
		return nil, err
	}
	defer readCloser.Close()
	stream := json.NewDecoder(readCloser)

	testIDMap := func(ids []string) map[string]struct{} {
		return nil
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
