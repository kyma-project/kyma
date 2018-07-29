package overrides

import (
	"bytes"
	"text/template"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
)

const globalsTplStr = `
global:
  tlsCrt: "{{.ClusterTLSCert}}"
  tlsKey: "{{.ClusterTLSKey}}"
  isLocalEnv: {{isLocal}}
  domainName: "{{.Domain}}"
  remoteEnvCa: "{{.RemoteEnvCa}}"
  remoteEnvCaKey: "{{.RemoteEnvCaKey}}"
  istio:
    tls:
      secretName: "istio-ingress-certs"
  etcdBackupABS:
    containerName: "{{.EtcdBackupABSContainerName}}"
`

// GetGlobalOverrides .
func GetGlobalOverrides(installationData *config.InstallationData) (string, error) {

	fmap := template.FuncMap{
		"isLocal": installationData.IsLocalInstallation,
	}

	tmpl, err := template.New("").Funcs(fmap).Parse(globalsTplStr)
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
