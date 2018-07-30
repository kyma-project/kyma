package toolkit

import (
	"github.com/kyma-project/kyma/components/installer/pkg/config"
)

// InstallationDataCreator .
type InstallationDataCreator struct {
	installationData config.InstallationData
}

// NewInstallationDataCreator return new instance of InstallationDataCreator
func NewInstallationDataCreator() *InstallationDataCreator {
	res := &InstallationDataCreator{
		installationData: config.InstallationData{},
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

// WithEmptyDomain sets Domain property in InstallationData to empty value
func (sc *InstallationDataCreator) WithEmptyDomain() *InstallationDataCreator {
	sc.installationData.Domain = ""

	return sc
}

// WithCert sets Cert and CertKey properties
func (sc *InstallationDataCreator) WithCert(cert, certKey string) *InstallationDataCreator {
	sc.installationData.ClusterTLSCert = cert
	sc.installationData.ClusterTLSKey = certKey
	return sc
}

// WithDomain sets Domain property in InstallationData
func (sc *InstallationDataCreator) WithDomain(domain string) *InstallationDataCreator {
	sc.installationData.Domain = domain

	return sc
}

// WithEmptyIP sets IP address property in InstallationData to empty value
func (sc *InstallationDataCreator) WithEmptyIP() *InstallationDataCreator {
	sc.installationData.ExternalIPAddress = ""

	return sc
}

// WithIP sets IP address in InstallationData
func (sc *InstallationDataCreator) WithIP(ipAddr string) *InstallationDataCreator {
	sc.installationData.ExternalIPAddress = ipAddr

	return sc
}

// WithRemoteEnvCa sets RemoteEnvCa property in InstallationData
func (sc *InstallationDataCreator) WithRemoteEnvCa(remoteEnvCa string) *InstallationDataCreator {
	sc.installationData.RemoteEnvCa = remoteEnvCa

	return sc
}

// WithRemoteEnvCaKey sets RemoteEnvCaKey property in InstallationData
func (sc *InstallationDataCreator) WithRemoteEnvCaKey(remoteEnvCaKey string) *InstallationDataCreator {
	sc.installationData.RemoteEnvCaKey = remoteEnvCaKey

	return sc
}

// WithRemoteEnvIP sets value for RemoteEnvIP property
func (sc *InstallationDataCreator) WithRemoteEnvIP(ipAddr string) *InstallationDataCreator {
	sc.installationData.RemoteEnvIP = ipAddr

	return sc
}

// WithUITestCredentials sets value for UITestUser and UITestPassword properties
func (sc *InstallationDataCreator) WithUITestCredentials(username, password string) *InstallationDataCreator {
	sc.installationData.UITestUser = username
	sc.installationData.UITestPassword = password

	return sc
}

// WithAdminGroup sets value for AdminGroup property
func (sc *InstallationDataCreator) WithAdminGroup(adminGroupName string) *InstallationDataCreator {
	sc.installationData.AdminGroup = adminGroupName

	return sc
}

// WithEtcdOperator sets value for EtcdOperator property
func (sc *InstallationDataCreator) WithEtcdOperator(enabled, storageAccount, storageKey string) *InstallationDataCreator {
	sc.installationData.EnableEtcdBackupOperator = enabled
	sc.installationData.EtcdBackupABSAccount = storageAccount
	sc.installationData.EtcdBackupABSKey = storageKey

	return sc
}

// WithEtcdBackupABSContainerName sets value for EtcdBackupABSContainerName property
func (sc *InstallationDataCreator) WithEtcdBackupABSContainerName(path string) *InstallationDataCreator {
	sc.installationData.EtcdBackupABSContainerName = path

	return sc
}

////////////////////////////////////////
// GetData returns InstallationData created by InstallationDataCreator
func (sc *InstallationDataCreator) GetData() config.InstallationData {
	sc.installationData.IsLocalInstallation = func() bool { return sc.installationData.ExternalIPAddress == "" }
	return sc.installationData
}
