package busola

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/busola-migrator/internal/config"
)

func TestBuildInitURL(t *testing.T) {
	// GIVEN

	appCfg := config.Config{
		BusolaURL:    "https://busola.url",
		KubeconfigID: "BCD86FA5-8D8E-4567-ACD9-511CCD4A7FF2",
	}
	urlRegexp := regexp.MustCompile(`^https://busola\.url/\?kubeconfigid=[0-9a-zA-Z-_]+$`)

	// WHEN
	res, err := BuildInitURL(appCfg)

	// THEN
	assert.NoError(t, err)
	assert.Regexp(t, urlRegexp, res)
}
