package config

import (
	"os"
	"strings"
)

// Configuration of non-secret values in installer
type installationConfig struct {
	ExternalPublicIP           string
	AzureBrokerTenantID        string
	AzureBrokerClientID        string
	AzureBrokerSubscriptionID  string
	AzureBrokerClientSecret    string
	ClusterTLSKey              string
	ClusterTLSCert             string
	RemoteEnvCa                string
	RemoteEnvCaKey             string
	EtcdBackupABSContainerName string
	ComponentsList             string
	IsLocalInstallation        bool
	VictorOpsApiKey            string
	VictorOpsRoutingKey        string
	SlackChannel               string
	SlackApiUrl                string
}

// GetInstallationConfig returns all non-secret installation parameters from the Installer environment variables
func GetInstallationConfig() *installationConfig {
	return &installationConfig{
		ExternalPublicIP:          os.Getenv("EXTERNAL_PUBLIC_IP"),
		AzureBrokerTenantID:       os.Getenv("AZURE_BROKER_TENANT_ID"),
		AzureBrokerClientID:       os.Getenv("AZURE_BROKER_CLIENT_ID"),
		AzureBrokerSubscriptionID: os.Getenv("AZURE_BROKER_SUBSCRIPTION_ID"),
		AzureBrokerClientSecret:   os.Getenv("AZURE_BROKER_CLIENT_SECRET"),
		ClusterTLSKey:             os.Getenv("TLS_KEY"),
		ClusterTLSCert:            os.Getenv("TLS_CERT"),
		RemoteEnvCa:               os.Getenv("REMOTE_ENV_CA"),
		RemoteEnvCaKey:            os.Getenv("REMOTE_ENV_CA_KEY"),
		ComponentsList:            os.Getenv("COMPONENT_LIST"),
		IsLocalInstallation:       isLocalInstallation(os.Getenv("IS_LOCAL_INSTALLATION")),
		VictorOpsApiKey:           os.Getenv("VICTOR_OPS_API_KEY_VALUE"),
		VictorOpsRoutingKey:       os.Getenv("VICTOR_OPS_ROUTING_KEY_VALUE"),
		SlackChannel:              os.Getenv("SLACK_CHANNEL_VALUE"),
		SlackApiUrl:               os.Getenv("SLACK_API_URL_VALUE"),
	}
}

func isLocalInstallation(value string) bool {
	return strings.ToLower(value) == "true"
}
