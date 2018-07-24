package overrides

import (
	"bytes"
	"text/template"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
)

const tplStr = `
azure-broker:
  enabled: true
  subscription_id: {{.AzSubscriptionID}}
  tenant_id: {{.AzTenantID}}
  client_id: {{.AzClientID}}
  client_secret: {{.AzClientSecret}}
`

type azureParams struct {
	AzTenantID       string
	AzClientID       string
	AzSubscriptionID string
	AzClientSecret   string
}

// EnableAzureBroker provides Azure parameters from Vault
func EnableAzureBroker(installationData *config.InstallationData) (string, error) {

	if !hasAzureParams(installationData) {
		return "", nil
	}

	azureParams, err := getAzureParams(installationData)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("").Parse(tplStr)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, azureParams)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func getAzureParams(installationData *config.InstallationData) (*azureParams, error) {

	azTenantID := installationData.AzureBrokerTenantID
	azClientID := installationData.AzureBrokerClientID
	azSubscriptionID := installationData.AzureBrokerSubscriptionID
	azClientSecret := installationData.AzureBrokerClientSecret

	ap := &azureParams{
		AzTenantID:       azTenantID,
		AzClientID:       azClientID,
		AzSubscriptionID: azSubscriptionID,
		AzClientSecret:   azClientSecret,
	}

	return ap, nil
}

func hasAzureParams(installationData *config.InstallationData) bool {
	return installationData.AzureBrokerTenantID != "" && installationData.AzureBrokerClientID != "" && installationData.AzureBrokerSubscriptionID != "" && installationData.AzureBrokerClientSecret != ""
}
