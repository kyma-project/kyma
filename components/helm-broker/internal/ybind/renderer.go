package ybind

import (
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/engine"
	"k8s.io/helm/pkg/proto/hapi/chart"
	rls "k8s.io/helm/pkg/proto/hapi/services"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

const (
	goTplEngine = "gotpl"
	bindFile    = "bindTmpl"
)

//go:generate mockery -name=chartGoTemplateRenderer -output=automock -outpkg=automock -case=underscore
type chartGoTemplateRenderer interface {
	Render(*chart.Chart, chartutil.Values) (map[string]string, error)
}

type toRenderValuesCaps func(*chart.Chart, *chart.Config, chartutil.ReleaseOptions, *chartutil.Capabilities) (chartutil.Values, error)

// Renderer purpose is to render helm template directives, like: {{ .Release.Namespace }}
type Renderer struct {
	renderEngine       chartGoTemplateRenderer
	toRenderValuesCaps toRenderValuesCaps
}

// NewRenderer creates new instance of Renderer.
func NewRenderer() *Renderer {
	return &Renderer{
		renderEngine:       engine.New(),
		toRenderValuesCaps: chartutil.ToRenderValuesCaps,
	}
}

// Render renders given bindTemplate in context of helm Chart by e.g. replacing directives like: {{ .Release.Namespace }}
func (r *Renderer) Render(bindTemplate internal.BundlePlanBindTemplate, resp *rls.InstallReleaseResponse) (RenderedBindYAML, error) {
	if err := r.validateInstallReleaseResponse(resp); err != nil {
		return nil, errors.Wrap(err, "while validating input")
	}

	ch := resp.Release.Chart
	options := r.createReleaseOptions(resp)
	chartCap := &chartutil.Capabilities{}

	valsToRender, err := r.toRenderValuesCaps(ch, resp.Release.Config, options, chartCap)
	if err != nil {
		return nil, errors.Wrap(err, "while merging values to render")
	}

	ch.Templates = append(ch.Templates, &chart.Template{Name: bindFile, Data: bindTemplate})

	files, err := r.renderEngine.Render(ch, valsToRender)
	if err != nil {
		return nil, errors.Wrap(err, "while rendering files")
	}

	rendered, exits := files[fmt.Sprintf("%s/%s", ch.Metadata.Name, bindFile)]
	if !exits {
		return nil, fmt.Errorf("%v file was not resolved after rendering", bindFile)
	}

	return RenderedBindYAML(rendered), nil
}

func (*Renderer) validateInstallReleaseResponse(resp *rls.InstallReleaseResponse) error {
	if resp == nil {
		return fmt.Errorf("input parameter 'InstallReleaseResponse' cannot be nil")
	}

	if resp.Release == nil {
		return fmt.Errorf("'Release' filed from 'InstallReleaseResponse' is missing")
	}

	if resp.Release.Info == nil {
		return fmt.Errorf("'Info' filed from 'InstallReleaseResponse' is missing")
	}

	ch := resp.Release.Chart
	if ch.Metadata.Engine != "" && ch.Metadata.Engine != goTplEngine {
		return fmt.Errorf("chart %q requested non-existent template engine %q", ch.Metadata.Name, ch.Metadata.Engine)
	}

	return nil
}

func (*Renderer) createReleaseOptions(resp *rls.InstallReleaseResponse) chartutil.ReleaseOptions {
	return chartutil.ReleaseOptions{
		Name:      resp.Release.Name,
		Time:      resp.Release.Info.LastDeployed,
		Namespace: resp.Release.Namespace,
		Revision:  int(resp.Release.Version),
		IsInstall: true,
	}
}
