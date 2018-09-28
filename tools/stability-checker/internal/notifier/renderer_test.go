package notifier

import (
	"testing"

	"github.com/kyma-project/kyma/tools/stability-checker/internal/summary"
	"github.com/stretchr/testify/require"
)

func TestSummaryIsRendered(t *testing.T) {
	r, err := NewTestRenderer()
	require.NoError(t, err)

	_, _, _, err = r.RenderTestSummary(RenderTestSummaryInput{
		ShowTestStats: true,
		TestStats: []summary.SpecificTestStats{
			{
				Name:      "test-kubeless",
				Failures:  978,
				Successes: 0,
			},
			{
				Name:      "test-remote-env",
				Successes: 1000,
				Failures:  0,
			},
		}})
	require.NoError(t, err)
}
