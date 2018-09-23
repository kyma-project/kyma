package summary

import (
	"regexp"

	"github.com/pkg/errors"
)

// OutputProcessor process output from executing tests and detects test successes and test failures.
type OutputProcessor struct {
	failRegexp    *regexp.Regexp
	successRegexp *regexp.Regexp
	aggregator    statsAggregator
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
		return nil, errors.New( "regexp indicating failed tests has to have one capturing group (test name)")
	}



	return &OutputProcessor{
		failRegexp:    failureR,
		successRegexp: successR,
		aggregator:    newStatsAggregator(),
	}, nil
}

// Process process test ouptut and stores results internally.
// This method can be called many times. To get results, use GetResults method.
func (p *OutputProcessor) Process(input []byte) error {
	if err := p.findSuccessIndicator(input); err != nil {
		return err
	}
	return p.findFailureIndicator(input)
}

func (p *OutputProcessor) findSuccessIndicator(input []byte) error {
	successSubmatches := p.successRegexp.FindAllSubmatch([]byte(input), -1)
	for _, sm := range successSubmatches {
		if len(sm) != 2 {
			return errors.New("incorrect regexp for test success occurrences")
		}
		testName := sm[1]
		p.aggregator.addTestResult(string(testName), true)
	}
	return nil
}

func (p *OutputProcessor) findFailureIndicator(input []byte) error {
	failedMatches := p.failRegexp.FindAllSubmatch([]byte(input), -1)
	for _, fm := range failedMatches {
		if len(fm) != 2 {
			return errors.New("incorrect regexp for test failure occurrences")
		}
		testName := fm[1]
		p.aggregator.addTestResult(string(testName), false)
	}

	return nil
}

// GetResults returns statistics how many times every test passed and failed.
func (p *OutputProcessor) GetResults() []SpecificTestStats {
	list := make([]SpecificTestStats, len(p.aggregator.m))
	i := 0
	for _, v := range p.aggregator.m {
		list[i] = *v
		i++
	}
	return list
}

type statsAggregator struct {
	m map[string]*SpecificTestStats
}

func newStatsAggregator() statsAggregator {
	return statsAggregator{
		m: make(map[string]*SpecificTestStats),
	}
}

func (sc *statsAggregator) addTestResult(testName string, result bool) {
	_, ex := sc.m[testName]
	if ex {
		sc.m[testName].add(result)
	} else {
		newTest := &SpecificTestStats{Name: testName}
		newTest.add(result)
		sc.m[testName] = newTest
	}
}

// SpecificTestStats aggregates information how many specific test ends up with success and failure.
type SpecificTestStats struct {
	Name      string
	Successes int
	Failures  int
}

func (b *SpecificTestStats) add(val bool) {
	if val {
		b.Successes++
	} else {
		b.Failures++
	}
}
