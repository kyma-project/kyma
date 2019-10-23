package kube_config

import (
	"io"
	"text/template"

	log "github.com/sirupsen/logrus"
)

var content = `
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: {{.CA}}
    server: {{.URL}}
  name: {{.ClusterName}}
contexts:
- context:
    cluster: {{.ClusterName}}
    {{- if .NS}}
    namespace: {{.NS}}
    {{- end}}
    user: OIDCUser
  name: {{.ClusterName}}
current-context: {{.ClusterName}}
kind: Config
preferences: {}
users:
- name: OIDCUser
  user:
    token: {{.Token}}
`

type CAValueProvider func() string

type KubeConfig struct {
	clusterName string
	url         string
	caProvider  CAValueProvider
	namespace   string
	tmpl        *template.Template
}

//NewKubeConfig return new instance of KubeConfig
//CAValueProvider is used to return current cluster CA data, it can change after cluster certificates renewal.
func NewKubeConfig(clusterName, url string, caProvider CAValueProvider, namespace string) *KubeConfig {

	tmpl := template.Must(template.New("kubeConfig").Parse(content))

	return &KubeConfig{
		clusterName: clusterName,
		url:         url,
		caProvider:  caProvider,
		namespace:   namespace,
		tmpl:        tmpl,
	}
}

type data struct {
	ClusterName string
	URL         string
	CA          string
	NS          string
	Token       string
}

func (c *KubeConfig) Generate(output io.Writer, token string) {

	log.Debug("Generating kube config...")

	d := data{
		ClusterName: c.clusterName,
		URL:         c.url,
		CA:          c.caProvider(),
		NS:          c.namespace,
		Token:       token,
	}

	c.tmpl.Execute(output, d)

	log.Debug("Kube config generated.")
}
