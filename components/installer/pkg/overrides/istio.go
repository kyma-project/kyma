package overrides

import (
	"bytes"
	"text/template"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
)

const istioTplStr = `
ingressgateway:
  service:
    externalPublicIp: {{.ExternalPublicIP}}
`

// GetIstioOverrides returns values overrides for istio ingressgateway
func GetIstioOverrides(installationData *config.InstallationData) (OverridesMap, error) {
	if hasIPAddress(installationData) == false {
		return OverridesMap{}, nil
	}

	tmpl, err := template.New("").Parse(istioTplStr)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	bErr := tmpl.Execute(buf, installationData)
	if bErr != nil {
		return nil, err
	}

	return ToMap(buf.String())
}

func hasIPAddress(installationData *config.InstallationData) bool {
	return installationData.ExternalPublicIP != ""
}
