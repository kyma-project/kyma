package summary

import (
	"regexp"

	"github.com/pkg/errors"
)

// OutputProcessor process output from executing tests and detects test successes and test failures.
type OutputProcessor struct {
	failRegexp    *regexp.Regexp
	successRegexp *regexp.Regexp
}

// NewOutputProcessor is a constructor for OutputProcessor.
// failRegexp and successRegexp defines indicator of failure and success. Both regexp has to define capturing group for a test name.
func NewOutputProcessor(failRegexp, successRegexp string) (*OutputProcessor, error) {
	successR, err := regexp.Compile(successRegexp)
	if err != nil {
		return nil, errors.Wrap(err, "while compiling regexp indicating successful tests")
	}

	if len(successR.SubexpNames()) != 2 {
		return nil, errors.New("regexp indicating successful tests has to have one capturing group (test name)")
	}

	failureR, err := regexp.Compile(failRegexp)
	if err != nil {
		return nil, errors.Wrap(err, "while compiling regexp indication failed tests")
	}

	if len(failureR.SubexpNames()) != 2 {
		return nil, errors.New("regexp indicating failed tests has to have one capturing group (test name)")
	}

	return &OutputProcessor{
		failRegexp:    failureR,
		successRegexp: successR,
	}, nil
}

// Process process test output and returns statistics for specific tests
func (p *OutputProcessor) Process(input []byte) (map[string]SpecificTestStats, error) {
	aggr := newStatsAggregator()
	if err := p.findSuccessIndicator(input, aggr); err != nil {
		return nil, err
	}
	if err := p.findFailureIndicator(input, aggr); err != nil {
		return nil, err
	}
	return aggr.ToMap(), nil

}

func (p *OutputProcessor) findSuccessIndicator(input []byte, aggr *statsAggregator) error {
	successSubmatches := p.successRegexp.FindAllSubmatch([]byte(input), -1)
	for _, sm := range successSubmatches {
		if len(sm) != 2 {
			return errors.New("incorrect regexp for test success occurrences")
		}
		testName := sm[1]
		aggr.AddTestResult(string(testName), true)
	}
	return nil
}

func (p *OutputProcessor) findFailureIndicator(input []byte, aggr *statsAggregator) error {
	failedMatches := p.failRegexp.FindAllSubmatch([]byte(input), -1)
	for _, fm := range failedMatches {
		if len(fm) != 2 {
			return errors.New("incorrect regexp for test failure occurrences")
		}
		testName := fm[1]
		aggr.AddTestResult(string(testName), false)
	}
	return nil
}
