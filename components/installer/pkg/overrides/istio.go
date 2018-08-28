package overrides

import (
	"bytes"
	"text/template"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
)

const istioTplStr = `
gateways:
  istio-ingressgateway:
    service:
      externalPublicIp: {{.ExternalPublicIP}}
`

// GetIstioOverrides returns values overrides for istio ingressgateway
func GetIstioOverrides(installationData *config.InstallationData, overrides Map) (Map, error) {
	if hasIPAddress(installationData) == false {
		return Map{}, nil
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

	istioOverrides, err := ToMap(buf.String())
	if err != nil {
		return nil, err
	}

	allOverrides := Map{}
	MergeMaps(allOverrides, overrides)
	MergeMaps(allOverrides, istioOverrides)

	return allOverrides, nil
}

func hasIPAddress(installationData *config.InstallationData) bool {
	return installationData.ExternalPublicIP != ""
}
