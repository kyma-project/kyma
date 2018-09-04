package overrides

import (
	"github.com/kyma-project/kyma/components/installer/pkg/config"
)

// GetCoreOverrides - returns values overrides for core installation basing on domain
func GetCoreOverrides(installationData *config.InstallationData, overrides Map) (Map, error) {
	return overrides, nil
}
