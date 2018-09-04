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
	ExternalPublicIP           string
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
	EtcdBackupABSContainerName string
	Components                 []v1alpha1.KymaComponent
	IsLocalInstallation        bool
	VictorOpsApiKey            string
	VictorOpsRoutingKey        string
	SlackChannel               string
	SlackApiUrl                string
}

// NewInstallationData .
func NewInstallationData(installation *v1alpha1.Installation, installationConfig *installationConfig) (*InstallationData, error) {

	ctx := InstallationContext{
		Name:      installation.Name,
		Namespace: installation.Namespace,
	}

	res := &InstallationData{
		Context:                    ctx,
		ExternalPublicIP:           installationConfig.ExternalPublicIP,
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
		EtcdBackupABSContainerName: installationConfig.EtcdBackupABSContainerName,
		Components:                 installation.Spec.Components,
		IsLocalInstallation:        installationConfig.IsLocalInstallation,
		VictorOpsApiKey:            installationConfig.VictorOpsApiKey,
		VictorOpsRoutingKey:        installationConfig.VictorOpsRoutingKey,
		SlackChannel:               installationConfig.SlackChannel,
		SlackApiUrl:                installationConfig.SlackApiUrl,
	}
	return res, nil
}
