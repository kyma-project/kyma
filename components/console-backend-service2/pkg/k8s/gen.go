// +build ignore

package main

import (
	"flag"
	"os"
	"strings"
	"text/template"
)

type Resources []string

func (r Resources) String() string {
	return strings.Join(r, ",")
}

func (r *Resources) Set(s string) error {
	*r = strings.Split(s, ",")
	return nil
}

type config struct {
	Group        string
	TypesPackage string
	Resources    Resources
}

func main() {
	cfg := config{}
	flag.StringVar(&cfg.Group, "group", "", "")
	flag.StringVar(&cfg.TypesPackage, "types-package", "", "")
	flag.Var(&cfg.Resources, "resources", "")
	err := flag.CommandLine.Parse(os.Args[2:])
	if err != nil {
		panic(err)
	}

	f, err := os.Create(strings.ToLower(cfg.Group) + ".go")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = packageTemplate.Execute(f, cfg)
	if err != nil {
		panic(err)
	}
}

var packageTemplate = template.Must(template.New("").Funcs(template.FuncMap{
	"lower": strings.ToLower,
}).Parse(`
package k8s

import (
	types "{{ .TypesPackage }}"
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/resource"
)

type {{ .Group }}Services struct {
{{- range .Resources }}
	{{ . }} *resource.Service
{{- end }}
}

func New{{ .Group }}Services(serviceFactory *resource.ServiceFactory) *{{ .Group }}Services {
	return &{{ .Group }}Services{
{{- range .Resources }}
		{{ . }}: serviceFactory.ForResource(types.SchemeGroupVersion.WithResource("{{ . | lower }}")),
{{- end }}
	}
}
`))
