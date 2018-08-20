package overrides

import (
	"bytes"
	"text/template"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
)

const globalsTplStr = `
global:
  tlsPEMCrt: "{{.ClusterTLSPEMCert}}"
  tlsCrt: "{{.ClusterTLSCert}}"
  tlsKey: "{{.ClusterTLSKey}}"
  isLocalEnv: {{.IsLocalInstallation}}
  domainName: "{{.Domain}}"
  remoteEnvCa: "{{.RemoteEnvCa}}"
  remoteEnvCaKey: "{{.RemoteEnvCaKey}}"
  istio:
    tls:
      secretName: "istio-ingress-certs"
  etcdBackupABS:
    containerName: "{{.EtcdBackupABSContainerName}}"
  alertTools:
    credentials:
      victorOps:
        routingkey: "{{ .VictorOpsRoutingKey }}"
        apikey: "{{ .VictorOpsApiKey }}"
      slack:
        channel: "{{ .SlackChannel }}"
        apiurl: "{{ .SlackApiUrl }}"
`

// GetGlobalOverrides .
func GetGlobalOverrides(installationData *config.InstallationData) (string, error) {

	tmpl, err := template.New("").Parse(globalsTplStr)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, installationData)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
