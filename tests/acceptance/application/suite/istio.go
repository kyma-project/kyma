package suite

import (
	"bytes"
	"text/template"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const (
	denierDefinition = `
apiVersion: "config.istio.io/v1alpha2"
kind: denier
metadata:
  name: gw
  namespace: {{.Namespace}}
spec:
  status:
    code: 7
    message: |
      Not allowed by istio denier
`
	checkNothingDefinition = `
apiVersion: "config.istio.io/v1alpha2"
kind: checknothing
metadata:
  name: gw
  namespace: {{.Namespace}}
spec:
`

	ruleDefinition = `
apiVersion: "config.istio.io/v1alpha2"
kind: rule
metadata:
  name: {{.Service}}
  namespace: {{.Namespace}}
spec:
  match: (destination.service.host == "{{.Service}}.{{.Namespace}}.svc.cluster.local") && (source.labels["{{.AccessLabel}}"] != "true")
  actions:
  - handler: gw.denier
    instances:
    - gw.checknothing
`
)

func (ts *TestSuite) createIstioResources() {
	var data = struct {
		Namespace   string
		AccessLabel string
		Service     string
	}{
		Namespace:   ts.namespace,
		AccessLabel: ts.accessLabel,
		Service:     ts.gatewaySvcName,
	}

	denierTmpl := template.Must(template.New("denier").Parse(denierDefinition))
	checkNothingTmpl := template.Must(template.New("checknothing").Parse(checkNothingDefinition))
	ruleTmpl := template.Must(template.New("rule").Parse(ruleDefinition))

	cp, err := dynamic.NewForConfig(ts.config)
	require.NoError(ts.t, err)

	for _, tmpl := range []*template.Template{denierTmpl, checkNothingTmpl, ruleTmpl} {
		obj := ts.unmarshal(data, tmpl)
		kind := obj["kind"].(string)

		dcl := cp.Resource(schema.GroupVersionResource{
			Resource: kind + "s",
			Version:  "v1alpha2",
			Group:    "config.istio.io",
		}).Namespace(ts.namespace)

		dcl.Create(&unstructured.Unstructured{Object: obj}, metav1.CreateOptions{})
	}
}

func (ts *TestSuite) unmarshal(data interface{}, tmpl *template.Template) map[string]interface{} {
	var obj map[string]interface{}
	var buffer bytes.Buffer
	err := tmpl.Execute(&buffer, &data)
	require.NoError(ts.t, err)
	err = yaml.Unmarshal(buffer.Bytes(), &obj)
	require.NoError(ts.t, err)

	return obj
}
