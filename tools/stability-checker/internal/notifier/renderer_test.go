package notifier

import (
	"fmt"
	"testing"

	"github.com/kyma-project/kyma/tools/stability-checker/internal/summary"
	"github.com/stretchr/testify/require"
)

func TestAbc(t *testing.T) {
	r, err := NewTestRenderer()
	require.NoError(t, err)

	a, b, c, err := r.RenderTestSummary(RenderTestSummaryInput{
		ShowTestStats: true,
		// TODO
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
	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(c)
}
