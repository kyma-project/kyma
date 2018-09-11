package overrides

import (
	"github.com/kyma-project/kyma/components/installer/pkg/config"
)

// GetGlobalOverrides .
func GetGlobalOverrides(installationData *config.InstallationData, overrides Map) (Map, error) {
	return overrides, nil
}
