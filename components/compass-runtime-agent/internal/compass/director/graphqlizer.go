package director

import (
	"bytes"
	"encoding/json"
	sprig "github.com/go-task/slim-sprig"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
	"reflect"
	"text/template"
)

type Graphqlizer struct{}

func (g *Graphqlizer) RuntimeUpdateInputToGQL(in graphql.RuntimeInput) (string, error) {
	return g.genericToGQL(in, `{
		name: "{{.Name}}",
		{{- if .Description }}
		description: "{{.Description}}",
		{{- end }}
		{{- if .Labels }}
		labels: {{ LabelsToGQL .Labels}},
		{{- end }}
		{{- if .StatusCondition }}
		statusCondition: {{ .StatusCondition }},
		{{- end }}
	}`)
}

// LabelsToGQL missing godoc
func (g *Graphqlizer) LabelsToGQL(in graphql.Labels) (string, error) {
	return g.marshal(in), nil
}

func (g *Graphqlizer) genericToGQL(obj interface{}, tmpl string) (string, error) {
	fm := sprig.TxtFuncMap()
	fm["marshal"] = g.marshal
	fm["LabelsToGQL"] = g.LabelsToGQL

	t, err := template.New("tmpl").Funcs(fm).Parse(tmpl)
	if err != nil {
		return "", errors.Wrapf(err, "while parsing template")
	}

	var b bytes.Buffer

	if err := t.Execute(&b, obj); err != nil {
		return "", errors.Wrap(err, "while executing template")
	}
	return b.String(), nil
}
func (g *Graphqlizer) marshal(obj interface{}) string {
	var out string

	val := reflect.ValueOf(obj)

	switch val.Kind() {
	case reflect.Map:
		s, err := g.genericToGQL(obj, `{ {{- range $k, $v := . }}{{ $k }}:{{ marshal $v }},{{ end -}} }`)
		if err != nil {
			return ""
		}
		out = s
	case reflect.Slice, reflect.Array:
		s, err := g.genericToGQL(obj, `[{{ range $i, $e := . }}{{ if $i }},{{ end }}{{ marshal $e }}{{ end }}]`)
		if err != nil {
			return ""
		}
		out = s
	default:
		marshalled, err := json.Marshal(obj)
		if err != nil {
			return ""
		}
		out = string(marshalled)
	}

	return out
}
