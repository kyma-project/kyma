package toolkit

import (
	"github.com/kyma-project/kyma/components/installer/pkg/config"
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

// WithEmptyAzureCredentials sets azure credentials to empty values
func (sc *InstallationDataCreator) WithEmptyAzureCredentials() *InstallationDataCreator {
	sc.installationData.AzureBrokerClientID = ""
	sc.installationData.AzureBrokerClientSecret = ""
	sc.installationData.AzureBrokerSubscriptionID = ""
	sc.installationData.AzureBrokerTenantID = ""

	return sc
}

// WithDummyAzureCredentials sets azure credentials in InstallationData to dummy values
func (sc *InstallationDataCreator) WithDummyAzureCredentials() *InstallationDataCreator {
	sc.installationData.AzureBrokerClientID = "37bb544f-8935-4a00-a934-3999577fb637"
	sc.installationData.AzureBrokerClientSecret = "ZGM3ZDlkYTgtZWMxMS00NTg4LTk5OGItOGU5YWJlNWUzYmE4DQo="
	sc.installationData.AzureBrokerSubscriptionID = "d5423a63-0ab6-4455-9efe-569c6e716625"
	sc.installationData.AzureBrokerTenantID = "7ffdff3c-daa6-420d-9cff-b04769031acf"

	return sc
}

// WithGeneric sets generic property in InstallationData
func (sc *InstallationDataCreator) WithGeneric(key, value string) *InstallationDataCreator {
	sc.genericOverrides[key] = value
	return sc
}

// WithLocalInstallation sets IsLocalInstallation poperty in InstallationData to true
func (sc *InstallationDataCreator) WithLocalInstallation() *InstallationDataCreator {
	sc.installationData.IsLocalInstallation = true

	return sc
}

////////////////////////////////////////
// GetData returns InstallationData created by InstallationDataCreator
func (sc *InstallationDataCreator) GetData() (config.InstallationData, map[string]string) {

	//We can't use types from overrides package here because of cyclic imports, we must use general types.
	return sc.installationData, sc.genericOverrides
}
