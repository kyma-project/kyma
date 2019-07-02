package toolkit

import (
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/config"
)

// InstallationDataCreator .
type InstallationDataCreator struct {
	installationData config.InstallationData
	genericOverrides map[string]string
}

// NewInstallationDataCreator return new instance of InstallationDataCreator
func NewInstallationDataCreator() *InstallationDataCreator {
	res := &InstallationDataCreator{
		installationData: config.InstallationData{},
		genericOverrides: map[string]string{},
	}

	return res
}

// WithGeneric sets generic property in InstallationData
func (sc *InstallationDataCreator) WithGeneric(key, value string) *InstallationDataCreator {
	sc.genericOverrides[key] = value
	return sc
}

// GetData returns InstallationData created by InstallationDataCreator
func (sc *InstallationDataCreator) GetData() (config.InstallationData, map[string]string) {

	//We can't use types from overrides package here because of cyclic imports, we must use general types.
	return sc.installationData, sc.genericOverrides
}
