package busola

import (
	"fmt"
	"net/url"

	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/busola-migrator/internal/config"
)

func BuildInitURL(appCfg config.Config) (string, error) {

	initURL, err := url.ParseRequestURI(fmt.Sprintf("%s/?kubeconfigID=%s", appCfg.BusolaURL, appCfg.KubeconfigID))
	if err != nil {
		return "", errors.Wrap(err, "while parsing Busola init url")
	}

	return initURL.String(), nil
}
