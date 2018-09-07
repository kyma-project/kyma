package config

import (
	"os"
	"strings"
)

// Configuration of non-secret values in installer
type installationConfig struct {
	IsLocalInstallation bool
}

// GetInstallationConfig returns all non-secret installation parameters from the Installer environment variables
func GetInstallationConfig() *installationConfig {
	return &installationConfig{
		IsLocalInstallation: isLocalInstallation(os.Getenv("IS_LOCAL_INSTALLATION")),
	}
}

func isLocalInstallation(value string) bool {
	return strings.ToLower(value) == "true"
}
