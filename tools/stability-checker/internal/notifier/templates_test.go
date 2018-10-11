package notifier

import (
	"github.com/stretchr/testify/require"
	"html/template"
	"io/ioutil"
	"testing"
)

func TestRenderHeader(t *testing.T) {
	// GIVEN
	tpl, err := template.New("header").Parse(header)
	require.NoError(t, err)
	// WHEN
	err = tpl.Execute(ioutil.Discard, RenderTestSummaryInput{})
	// THEN
	require.NoError(t, err)
}
