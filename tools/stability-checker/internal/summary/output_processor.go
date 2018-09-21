package summary

import (
	"regexp"

	"github.com/pkg/errors"
)

// OutputTestProcessor ...
type OutputTestProcessor struct {
	FailRegexp    *regexp.Regexp
	SuccessRegexp *regexp.Regexp
	aggregator    statsAggregator
}

// NewOutputTestProcessor ...
func NewOutputTestProcessor(failRegexp, successRegexp string) (*OutputTestProcessor, error) {
	successR, err := regexp.Compile(successRegexp)
	if err != nil {
		return nil, errors.Wrap(err, "while compiling regexp indicating successful tests")
	}

	if len(successR.SubexpNames()) != 2 {
		return nil, errors.Wrap(err, "regexp indicating successful tests has to have one capturing group (test name)")
	}

	failureR, err := regexp.Compile(failRegexp)
	if len(failureR.SubexpNames()) != 2 {
		return nil, errors.Wrap(err, "regexp indicating failed tests has to have one capturing group (test name)")
	}

	if err != nil {
		return nil, errors.Wrap(err, "while compiling regexp indication failed tests")
	}

	return &OutputTestProcessor{
		FailRegexp:    failureR,
		SuccessRegexp: successR,
		aggregator:    newStatsAggregator(),
	}, nil
}

// Process just do
func (p *OutputTestProcessor) Process(input []byte) error {
	if err := p.findSuccessIndicator(input); err != nil {
		return err
	}
	return p.findFailureIndicator(input)
}

func (p *OutputTestProcessor) findSuccessIndicator(input []byte) error {
	successSubmatches := p.SuccessRegexp.FindAllSubmatch([]byte(input), -1)
	for _, sm := range successSubmatches {
		if len(sm) != 2 {
			return errors.New("some error")
		}
		testName := sm[1]
		p.aggregator.Add(string(testName), true)
	}
	return nil
}

func (p *OutputTestProcessor) findFailureIndicator(input []byte) error {
	failedMatches := p.FailRegexp.FindAllSubmatch([]byte(input), -1)
	for _, fm := range failedMatches {
		if len(fm) != 2 {
			return errors.New("some error")
		}
		testName := fm[1]
		p.aggregator.Add(string(testName), false)
	}

	return nil
}

// GetResults ...
func (p *OutputTestProcessor) GetResults() []SpecificTestStats {
	list := make([]SpecificTestStats, len(p.aggregator.m))
	i := 0
	for _, v := range p.aggregator.m {
		list[i] = v
		i++
	}
	return list
}

type statsAggregator struct {
	m map[string]SpecificTestStats
}

func newStatsAggregator() statsAggregator {
	return statsAggregator{
		m: make(map[string]SpecificTestStats),
	}
}

// Add adds
func (sc *statsAggregator) Add(testName string, result bool) {
	curr, ex := sc.m[testName]
	if ex {
		curr.Add(result)
	} else {
		b := SpecificTestStats{Name: testName}
		b.Add(result)
		sc.m[testName] = b
	}
}

// SpecificTestStats ...
type SpecificTestStats struct {
	Name      string
	Successes int
	Failures  int
}

// Add adds ...
func (b *SpecificTestStats) Add(val bool) {
	if val {
		b.Successes++
	} else {
		b.Failures++
	}
}
