package overrides

import (
	"github.com/kyma-project/kyma/components/installer/pkg/config"
)

// GetIstioOverrides returns values overrides for istio ingressgateway
func GetIstioOverrides(installationData *config.InstallationData, overrides Map) (Map, error) {
	return overrides, nil
}
