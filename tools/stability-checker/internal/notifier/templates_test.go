package notifier

import (
	"html/template"
	"io/ioutil"
	"testing"

	"github.com/kyma-project/kyma/tools/stability-checker/internal"
	"github.com/kyma-project/kyma/tools/stability-checker/internal/summary"
	"github.com/stretchr/testify/require"
)

func TestRenderTemplates(t *testing.T) {
	toTest := map[string]string{
		"header": header,
		"body":   body,
		"footer": footer,
	}

	for tName, tpl := range toTest {
		// GIVEN
		tpl, err := template.New(tName).Parse(tpl)
		require.NoError(t, err)
		// WHEN
		err = tpl.Execute(ioutil.Discard, RenderTestSummaryInput{
			TotalTestsCnt: 10,
			FailedExecutions: []internal.ExecutionStatus{
				{ID: "123", Pass: false},
			},
			ShowTestStats: true,
			TestStats: []summary.SpecificTestStats{
				{
					Name:      "test-kubless",
					Failures:  123,
					Successes: 2,
				},
				{
					Name:      "test-very-long-name",
					Failures:  0,
					Successes: 55,
				},
			},
		})
		// THEN
		require.NoError(t, err)
	}
}
