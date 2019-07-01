package azurevault

import (
	"strings"
)

// SecretProviderInterface exposes functions to interact with azure-vault client
type SecretProviderInterface interface {
	GetSecret(secretName string) (string, error)
	GetCertificate(certName string) (string, error)
	GetKey(keyName string) (string, error)
}

// SecretProvider provides functions to interact with azure-vault client
// Implements SecretProviderInterface
type SecretProvider struct {
	client azureKeyvaultInterface
}

// NewSecretProvider .
func NewSecretProvider(azureClientID string, azureClientSecret string, vaultURL string) *SecretProvider {
	return &SecretProvider{
		client: getAzureKeyvaultClient(azureClientID, azureClientSecret, vaultURL),
	}
}

// GetSecret parses the object returned by azure-vault client and returns its value converted to string
func (sp *SecretProvider) GetSecret(secretName string) (string, error) {

	secretBundle, err := sp.client.GetSecret(secretName)
	if err != nil {
		return "", err
	}

	secret := removeNewLines(useString(secretBundle.Value))
	return secret, nil
}

// GetCertificate .
func (sp *SecretProvider) GetCertificate(certName string) (string, error) {
	return "", nil //todo
}

// GetKey .
func (sp *SecretProvider) GetKey(keyName string) (string, error) {
	return "", nil //todo
}

func useString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// We do that for two reasons:
// - Multiline string values can be used in yaml files with special syntax, but we don't support it (our templates are static).
// - K8s Helm client converts newline characters to spaces (in overrides), which breaks yaml deployments for secretes (base64 tokens contain spaces).
func removeNewLines(s string) string {
	return strings.Replace(s, "\n", "", -1)
}
