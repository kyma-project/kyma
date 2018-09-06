package config

import (
	"os"
	"strings"
)

// Configuration of non-secret values in installer
type installationConfig struct {
	AzureBrokerTenantID       string
	AzureBrokerClientID       string
	AzureBrokerSubscriptionID string
	AzureBrokerClientSecret   string
	ComponentsList            string
	IsLocalInstallation       bool
}

// GetInstallationConfig returns all non-secret installation parameters from the Installer environment variables
func GetInstallationConfig() *installationConfig {
	return &installationConfig{
		AzureBrokerTenantID:       os.Getenv("AZURE_BROKER_TENANT_ID"),
		AzureBrokerClientID:       os.Getenv("AZURE_BROKER_CLIENT_ID"),
		AzureBrokerSubscriptionID: os.Getenv("AZURE_BROKER_SUBSCRIPTION_ID"),
		AzureBrokerClientSecret:   os.Getenv("AZURE_BROKER_CLIENT_SECRET"),
		ComponentsList:            os.Getenv("COMPONENT_LIST"),
		IsLocalInstallation:       isLocalInstallation(os.Getenv("IS_LOCAL_INSTALLATION")),
	}
}

func isLocalInstallation(value string) bool {
	return strings.ToLower(value) == "true"
}
