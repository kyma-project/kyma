package overrides

import (
	"bytes"
	"text/template"

	"github.com/kyma-project/kyma/components/installer/pkg/config"
)

const coreTplStr = `
configurations-generator:
  kubeConfig:
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
func GetCoreOverrides(installationData *config.InstallationData, overrides Map) (Map, error) {

	tmpl, err := template.New("").Parse(coreTplStr)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, installationData)
	if err != nil {
		return nil, err
	}

	coreOverrides, err := ToMap(buf.String())
	if err != nil {
		return nil, err
	}

	allOverrides := Map{}
	MergeMaps(allOverrides, overrides)
	MergeMaps(allOverrides, coreOverrides)

	if hasValidDomain(allOverrides) == false {
		return Map{}, nil
	}

	return allOverrides, nil
}

func hasValidDomain(m Map) bool {

	res, found := FindOverrideValue(m, "global.domainName")

	if !found {
		return false
	}

	return res != ""
}
