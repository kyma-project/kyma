package config

import (
	"github.com/kyma-project/kyma/components/installer/pkg/apis/installer/v1alpha1"
)

// InstallationContext describes properties of K8S Installation object that triggers installation process
type InstallationContext struct {
	Name      string
	Namespace string
}

// InstallationData describes all installation attributes
type InstallationData struct {
	Context                    InstallationContext
	ExternalIPAddress          string
	Domain                     string
	KymaVersion                string
	URL                        string
	AzureBrokerTenantID        string
	AzureBrokerClientID        string
	AzureBrokerSubscriptionID  string
	AzureBrokerClientSecret    string
	ClusterTLSKey              string
	ClusterTLSCert             string
	RemoteEnvCa                string
	RemoteEnvCaKey             string
	RemoteEnvIP                string
	K8sApiserverURL            string
	K8sApiserverCa             string
	UITestUser                 string
	UITestPassword             string
	AdminGroup                 string
	EtcdBackupABSContainerName string
	EnableEtcdBackupOperator   string
	EtcdBackupABSAccount       string
	EtcdBackupABSKey           string
	IsLocalInstallation        func() bool
}

// NewInstallationData .
func NewInstallationData(installation *v1alpha1.Installation, installationConfig *installationConfig) (*InstallationData, error) {

	isLocalInstallationFunc := func() bool {
		return installationConfig.ExternalIPAddress == ""
	}

	ctx := InstallationContext{
		Name:      installation.Name,
		Namespace: installation.Namespace,
	}

	res := &InstallationData{
		Context:                    ctx,
		ExternalIPAddress:          installationConfig.ExternalIPAddress,
		Domain:                     installationConfig.Domain,
		KymaVersion:                installation.Spec.KymaVersion,
		URL:                        installation.Spec.URL,
		AzureBrokerTenantID:        installationConfig.AzureBrokerTenantID,
		AzureBrokerClientID:        installationConfig.AzureBrokerClientID,
		AzureBrokerSubscriptionID:  installationConfig.AzureBrokerSubscriptionID,
		AzureBrokerClientSecret:    installationConfig.AzureBrokerClientSecret,
		ClusterTLSKey:              installationConfig.ClusterTLSKey,
		ClusterTLSCert:             installationConfig.ClusterTLSCert,
		RemoteEnvCa:                installationConfig.RemoteEnvCa,
		RemoteEnvCaKey:             installationConfig.RemoteEnvCaKey,
		RemoteEnvIP:                installationConfig.RemoteEnvIP,
		K8sApiserverURL:            installationConfig.K8sApiserverUrl,
		K8sApiserverCa:             installationConfig.K8sApiserverCa,
		UITestUser:                 installationConfig.UITestUser,
		UITestPassword:             installationConfig.UITestPassword,
		AdminGroup:                 installationConfig.AdminGroup,
		EtcdBackupABSContainerName: installationConfig.EtcdBackupABSContainerName,
		EnableEtcdBackupOperator:   installationConfig.EnableEtcdBackupOperator,
		EtcdBackupABSAccount:       installationConfig.EtcdBackupABSAccount,
		EtcdBackupABSKey:           installationConfig.EtcdBackupABSKey,
		IsLocalInstallation:        isLocalInstallationFunc,
	}
	return res, nil
}
