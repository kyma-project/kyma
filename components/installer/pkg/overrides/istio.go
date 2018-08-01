package overrides

import (
	"bytes"
	"text/template"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
)

const istioTplStr = `
ingressgateway:
  service:
    externalPublicIp: {{.ExternalIPAddress}}
`

// GetIstioOverrides returns values overrides for istio ingressgateway
func GetIstioOverrides(installationData *config.InstallationData) (string, error) {
	if hasIPAddress(installationData) == false {
		return "", nil
	}

	tmpl, err := template.New("").Parse(istioTplStr)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	bErr := tmpl.Execute(buf, installationData)
	if bErr != nil {
		return "", err
	}

	return buf.String(), nil
}

func hasIPAddress(installationData *config.InstallationData) bool {
	return installationData.ExternalIPAddress != ""
}
