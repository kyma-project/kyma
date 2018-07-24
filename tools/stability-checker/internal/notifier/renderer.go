package notifier

import (
	"bytes"
	"text/template"
	"time"

	"github.com/kyma-project/kyma/tools/stability-checker/internal"
	"github.com/pkg/errors"
)

// TestRenderer renders test summary
type TestRenderer struct {
	headerReportTmpl *template.Template
	bodyReportTmpl   *template.Template
	footerReportTmpl *template.Template
}

// NewTestRenderer returns new instance of TestRenderer
func NewTestRenderer() (*TestRenderer, error) {
	headerReportTmpl, err := template.New("header").Parse(header)
	if err != nil {
		return nil, errors.Wrapf(err, "while parsing header template")
	}

	bodyReportTmpl, err := template.New("body").Parse(body)
	if err != nil {
		return nil, errors.Wrapf(err, "while parsing body template")
	}

	footerReportTmpl, err := template.New("footer").Parse(footer)
	if err != nil {
		return nil, errors.Wrapf(err, "while parsing footer template")
	}

	return &TestRenderer{
		headerReportTmpl: headerReportTmpl,
		bodyReportTmpl:   bodyReportTmpl,
		footerReportTmpl: footerReportTmpl,
	}, nil
}

// RenderTestSummaryInput holds input parameters required to render test summary
type RenderTestSummaryInput struct {
	TestResultWindowTime time.Duration
	TotalTestsCnt        int
	FailedTests          []internal.TestStatus
	TestRunnerInfo       TestRunnerInfo
}

// TestRunnerInfo describes test runner in kubernetes system
type TestRunnerInfo struct {
	PodName   string
	Namespace string
}

// RenderTestSummary returns header and body summary of given tests
func (s *TestRenderer) RenderTestSummary(in RenderTestSummaryInput) (string, string, string, error) {
	header := &bytes.Buffer{}
	if err := s.headerReportTmpl.Execute(header, in); err != nil {
		return "", "", "", errors.Wrapf(err, "while executing header template")
	}

	body := &bytes.Buffer{}
	if err := s.bodyReportTmpl.Execute(body, in); err != nil {
		return "", "", "", errors.Wrapf(err, "while executing body template")
	}

	footer := &bytes.Buffer{}
	if err := s.footerReportTmpl.Execute(footer, in); err != nil {
		return "", "", "", errors.Wrapf(err, "while executing footer template")
	}

	return header.String(), body.String(), footer.String(), nil
}
