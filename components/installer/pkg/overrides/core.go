package overrides

import (
	"bytes"
	"text/template"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
)

const coreTplStr = `
nginx-ingress:
  controller:
    service:
      loadBalancerIP: {{.RemoteEnvIP}}
configurations-generator:
  kubeConfig:
    clusterName: {{.Domain}}
    url: {{.K8sApiserverURL}}
    ca: {{.K8sApiserverCa}}
cluster-users:
  users:
    adminGroup: {{.AdminGroup}}
test:
  auth:
    username: "{{.UITestUser}}"
    password: "{{.UITestPassword}}"
etcd-operator:
  backupOperator:
    enabled: "{{.EnableEtcdBackupOperator}}"
    abs:
      storageAccount: "{{.EtcdBackupABSAccount}}"
      storageKey: "{{.EtcdBackupABSKey}}"
`

// GetCoreOverrides - returns values overrides for core installation basing on domain
func GetCoreOverrides(installationData *config.InstallationData) (Map, error) {
	if hasValidDomain(installationData) == false {
		return Map{}, nil
	}

	tmpl, err := template.New("").Parse(coreTplStr)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, installationData)
	if err != nil {
		return nil, err
	}

	return ToMap(buf.String())
}

func hasValidDomain(installationData *config.InstallationData) bool {
	return installationData.Domain != ""
}
