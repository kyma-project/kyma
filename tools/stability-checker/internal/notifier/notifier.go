package notifier

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/kyma-project/kyma/tools/stability-checker/internal/summary"

	"github.com/kyma-project/kyma/tools/stability-checker/internal"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SlackNotifier sends notification about test result to Slack channel.
type SlackNotifier struct {
	channelID            string
	cfgMapName           string
	cfgMapClient         configMapClient
	testResultWindowTime time.Duration
	log                  logrus.FieldLogger
	slack                slackClient
	testRenderer         testRenderer
	summarizer           summarizer
	testRunnerPodName    string
	testRunnerNamespace  string
	showTestStats        bool
}

type (
	slackClient interface {
		Send(header, body, footer, color string) error
	}
	testRenderer interface {
		RenderTestSummary(in RenderTestSummaryInput) (string, string, string, error)
	}
	summarizer interface {
		GetTestSummaryForExecutions(testIDs []string) ([]summary.SpecificTestStats, error)
	}
)

// New returns new instance of SlackNotifier
func New(
	slack slackClient,
	testRenderer testRenderer,
	cfgMapClient configMapClient,
	cfgMapName string,
	resultWindowTime time.Duration,
	testRunnerPodName, testRunnerNamespace string,
	log logrus.FieldLogger) *SlackNotifier {

	return &SlackNotifier{
		log:                  log,
		cfgMapName:           cfgMapName,
		cfgMapClient:         cfgMapClient,
		testResultWindowTime: resultWindowTime,
		slack:                slack,
		testRenderer:         testRenderer,
		testRunnerPodName:    testRunnerPodName,
		testRunnerNamespace:  testRunnerNamespace,
		showTestStats:        false,
	}
}

// Run sends in a loop notification about test result to Slack channel
func (s *SlackNotifier) Run(ctx context.Context) {
	for {
		if canceled := s.delay(ctx); canceled {
			return
		}

		cfg, err := s.cfgMapClient.Get(s.cfgMapName, metaV1.GetOptions{})
		if err != nil {
			s.log.Errorf("Cannot get ConfigMap %s, got error: %v", s.cfgMapName, err)
			continue
		}

		execResults := s.getTestExecutionFromTimeWindow(cfg)
		failedTests := s.filterFailingTests(execResults)

		var testStats []summary.SpecificTestStats
		if s.showTestStats {
			execIDs := make([]string, 0)
			for _, res := range execResults {
				execIDs = append(execIDs, res.ID)
			}
			testStats, err = s.summarizer.GetTestSummaryForExecutions(execIDs)
			if err != nil {
				s.log.Errorf("Cannot get test summary for execution IDs [%v], got error: %v", execIDs, err)
				continue
			}
		}

		header, body, footer, err := s.testRenderer.RenderTestSummary(RenderTestSummaryInput{
			FailedExecutions:     failedTests,
			TotalTestsCnt:        len(execResults),
			TestResultWindowTime: s.testResultWindowTime,
			TestRunnerInfo: TestRunnerInfo{
				PodName:   s.testRunnerPodName,
				Namespace: s.testRunnerNamespace,
			},
			ShowTestStats: s.showTestStats,
			TestStats:     testStats,
		})
		if err != nil {
			s.log.Errorf("Cannot render test summary, got error: %v", err)
			continue
		}

		if err := s.slack.Send(header, body, footer, color(failedTests)); err != nil {
			s.log.Errorf("Cannot send test summary, got error: %v", err)
		} else {
			s.log.Info("Sent slack notification successfully")
		}
	}
}

func (s *SlackNotifier) getTestExecutionFromTimeWindow(cfgMap *v1.ConfigMap) []internal.ExecutionStatus {
	var (
		lastSendTime = time.Now().Add(-s.testResultWindowTime)
		executions   []internal.ExecutionStatus
	)

	for k, v := range cfgMap.Data {
		testTime, err := s.parseToTime(k)
		if err != nil {
			s.log.Errorf("Cannot get test execution time, got error: %s", err)
			continue
		}

		if testTime.Before(lastSendTime) {
			continue
		}

		var exec internal.ExecutionStatus
		if err := json.Unmarshal([]byte(v), &exec); err != nil {
			s.log.Errorf("Cannot unmarshal execution results for entry with key %s, got error: %v", k, err)
			continue
		}

		executions = append(executions, exec)
	}

	return executions
}

func (s *SlackNotifier) parseToTime(timestamp string) (time.Time, error) {
	i, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return time.Time{}, errors.Wrapf(err, "while parsing %q to int type", timestamp)
	}
	tm := time.Unix(i, 0)

	return tm, nil
}

func (s *SlackNotifier) delay(ctx context.Context) bool {
	s.log.Debugf("Delay slack send for %v", s.testResultWindowTime)
	select {
	case <-ctx.Done():
		s.log.Debugf("Shutdown slack notifier because of %v", ctx.Err())
		return true
	case <-time.After(s.testResultWindowTime):
	}

	return false
}
func (s *SlackNotifier) filterFailingTests(testStatuses []internal.ExecutionStatus) (failed []internal.ExecutionStatus) {
	for _, test := range testStatuses {
		if !test.Pass {
			failed = append(failed, test)
		}
	}
	return
}

// WithSummarizer enables summarizer functionality for Slack Notifier
func (s *SlackNotifier) WithSummarizer(summarizer *summary.Service) {
	s.showTestStats = true
	s.summarizer = summarizer
}

func color(failed []internal.ExecutionStatus) string {
	var (
		red   = "#d92626"
		green = "#36a64f"
	)

	if len(failed) != 0 {
		return red
	}

	return green
}
