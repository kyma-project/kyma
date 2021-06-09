package busola

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/client-go/rest"

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

func BuildInitURL(appCfg config.Config, kubeConfig *rest.Config) (string, error) {
	if kubeConfig == nil {
		return "", errors.New("Kubeconfig not found")
	}

	host, err := url.Parse(kubeConfig.Host)
	if err != nil {
		return "", errors.Wrap(err, "while parsing apiserver host")
	}

	initString := fmt.Sprintf(initTemplate,
		getClusterName(host.Host),
		fmt.Sprintf("https://%s", host.Hostname()),
		base64.StdEncoding.EncodeToString(kubeConfig.CAData),
		appCfg.OIDC.IssuerURL,
		appCfg.OIDC.ClientID,
		appCfg.OIDC.Scope,
		appCfg.OIDC.UsePKCE,
	)
	encodedInitString, err := encodeInitString(initString)
	if err != nil {
		return "", errors.Wrap(err, "while encoding Busola init string payload")
	}

	initURL, err := url.ParseRequestURI(fmt.Sprintf("%s/?init=%s", appCfg.BusolaURL, encodedInitString))
	if err != nil {
		return "", errors.Wrap(err, "while parsing Busola init url")
	}

	return initURL.String(), nil
}

func getClusterName(host string) string {
	split := strings.Split(host, ".")
	if split[0] == "api" {
		return split[1]
	}
	return split[0]
}
