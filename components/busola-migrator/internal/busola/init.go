package busola

import (
	"fmt"
	"net/url"

	"github.com/pkg/errors"

	"github.com/kyma-project/kyma/components/busola-migrator/internal/config"
)

const initTemplate string = `{
 "kubeconfig": {
	  "apiVersion": "v1",
	  "kind": "Config",
	  "clusters": [
		{
		  "name": "%s",
		  "cluster": {
			"server": "%s",
			"certificate-authority-data": "%s"
		  }
		}
	  ],
	  "contexts": [
		{
		  "name": "%[1]s",
		  "context": {
			"cluster": "%[1]s",
			"user": "%[1]s-token"
		  }
		}
	  ],
	  "current-context": "%[1]s",
	  "users": [
		{
		  "name": "%[1]s-token",
		  "user": {
			"token": ""
		  }
		}
	  ]
	},
 "config": {
   "auth": {
     "issuerUrl": "%[4]s",
     "clientId": "%s",
     "scope": "%s",
     "usePKCE": %t
   },
	"hiddenNamespaces": [
     "istio-system",
     "knative-eventing",
     "knative-serving",
     "kube-public",
     "kube-system",
     "kyma-backup",
     "kyma-installer",
     "kyma-integration",
     "kyma-system",
     "natss",
     "kube-node-lease",
     "kubernetes-dashboard",
     "serverless-system"
   ],
   "navigation": {
     "disabledNodes": [],
     "externalNodes": []
   },
   "modules": {}
 }
}`

func BuildInitURL(appCfg config.Config) (string, error) {

	initURL, err := url.ParseRequestURI(fmt.Sprintf("%s/?kubeconfigID=%s", appCfg.BusolaURL, appCfg.KubeconfigID))
	if err != nil {
		return "", errors.Wrap(err, "while parsing Busola init url")
	}

	return initURL.String(), nil
}
