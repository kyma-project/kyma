package ybind_test

import (
	"errors"
	"fmt"
	"testing"

	google_protobuf "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/kyma-project/kyma/components/helm-broker/internal/ybind"
	"github.com/kyma-project/kyma/components/helm-broker/internal/ybind/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"
	hapi_release5 "k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/proto/hapi/services"
)

func TestRenderSuccess(t *testing.T) {
	// given
	fixResp := fixInstallReleaseResponse(fixChart())
	fixRenderOutFiles := map[string]string{
		fmt.Sprintf("%s/%s", fixChart().Metadata.Name, "bindTmpl"): "rendered-content",
	}
	tplToRender := internal.BundlePlanBindTemplate("template-body-to-render")

	engineRenderMock := &automock.ChartGoTemplateRenderer{}
	defer engineRenderMock.AssertExpectations(t)
	engineRenderMock.On("Render", mock.MatchedBy(chartWithTpl(t, tplToRender)), fixChartutilValues()).
		Return(fixRenderOutFiles, nil)

	toRenderFake := toRenderValuesFake{t}.WithInputAssertion(fixChart(), fixResp)
	renderer := ybind.NewRendererWithDeps(engineRenderMock, toRenderFake)

	// when
	out, err := renderer.Render(tplToRender, fixResp)

	// then
	require.NoError(t, err)
	assert.EqualValues(t, "rendered-content", out)
}

func TestRenderFailureOnInputParamValidation(t *testing.T) {
	for tn, tc := range map[string]struct {
		expErrMsg string
		givenResp *services.InstallReleaseResponse
	}{
		"response is nil": {
			expErrMsg: "input parameter 'InstallReleaseResponse' cannot be nil",
			givenResp: nil,
		},
		"missing Release filed": {
			expErrMsg: "'Release' filed from 'InstallReleaseResponse' is missing",
			givenResp: func() *services.InstallReleaseResponse {
				malformedResp := fixInstallReleaseResponse(fixChart())
				malformedResp.Release = nil
				return malformedResp
			}(),
		},
		"missing Info filed": {
			expErrMsg: "'Info' filed from 'InstallReleaseResponse' is missing",
			givenResp: func() *services.InstallReleaseResponse {
				malformedResp := fixInstallReleaseResponse(fixChart())
				malformedResp.Release.Info = nil
				return malformedResp
			}(),
		},
		"unsupported render engine": {
			expErrMsg: "chart \"test-chart\" requested non-existent template engine \"osm-engine\"",
			givenResp: func() *services.InstallReleaseResponse {
				malformedResp := fixInstallReleaseResponse(fixChart())
				malformedResp.Release.Chart.Metadata.Engine = "osm-engine"
				return malformedResp
			}(),
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			tplToRender := internal.BundlePlanBindTemplate("template-body-to-render")
			renderer := ybind.NewRendererWithDeps(nil, nil)

			// when
			out, err := renderer.Render(tplToRender, tc.givenResp)

			// then
			require.EqualError(t, err, fmt.Sprintf("while validating input: %s", tc.expErrMsg))
			assert.Nil(t, out)
		})
	}
}

func TestRenderFailureOnCreatingToRenderValues(t *testing.T) {
	// given
	fixErr := errors.New("fix err")
	fixResp := fixInstallReleaseResponse(fixChart())
	tplToRender := internal.BundlePlanBindTemplate("template-body-to-render")

	toRenderFake := toRenderValuesFake{t}.WithForcedError(fixErr)
	renderer := ybind.NewRendererWithDeps(nil, toRenderFake)

	// when
	out, err := renderer.Render(tplToRender, fixResp)

	// then
	require.EqualError(t, err, "while merging values to render: fix err")
	assert.Nil(t, out)
}

func TestRenderFailureOnEngineRender(t *testing.T) {
	// given
	fixResp := fixInstallReleaseResponse(fixChart())
	fixErr := errors.New("fix err")
	tplToRender := internal.BundlePlanBindTemplate("template-body-to-render")

	toRenderFake := toRenderValuesFake{t}.WithInputAssertion(fixChart(), fixResp)

	engineRenderMock := &automock.ChartGoTemplateRenderer{}
	defer engineRenderMock.AssertExpectations(t)
	engineRenderMock.On("Render", mock.MatchedBy(chartWithTpl(t, tplToRender)), fixChartutilValues()).
		Return(nil, fixErr)

	renderer := ybind.NewRendererWithDeps(engineRenderMock, toRenderFake)

	// when
	out, err := renderer.Render(tplToRender, fixResp)

	// then
	assert.EqualError(t, err, fmt.Sprintf("while rendering files: %s", fixErr))
	assert.Nil(t, out)
}

func TestRenderFailureOnExtractingResolveBindFile(t *testing.T) {
	// given
	fixResp := fixInstallReleaseResponse(fixChart())
	tplToRender := internal.BundlePlanBindTemplate("template-body-to-render")

	engineRenderMock := &automock.ChartGoTemplateRenderer{}
	defer engineRenderMock.AssertExpectations(t)
	engineRenderMock.On("Render", mock.MatchedBy(chartWithTpl(t, tplToRender)), fixChartutilValues()).
		Return(map[string]string{}, nil)

	toRenderFake := toRenderValuesFake{t}.WithInputAssertion(fixChart(), fixResp)
	renderer := ybind.NewRendererWithDeps(engineRenderMock, toRenderFake)

	// when
	out, err := renderer.Render(tplToRender, fixResp)

	// then
	assert.EqualError(t, err, "bindTmpl file was not resolved after rendering")
	assert.Nil(t, out)
}

func chartWithTpl(t *testing.T, expTpl internal.BundlePlanBindTemplate) func(*chart.Chart) bool {
	return func(ch *chart.Chart) bool {
		assert.Contains(t, ch.Templates, &chart.Template{Name: "bindTmpl", Data: expTpl})
		return true
	}
}

type toRenderValuesFake struct {
	t *testing.T
}

func (r toRenderValuesFake) WithInputAssertion(expChrt chart.Chart, expResp *services.InstallReleaseResponse) func(*chart.Chart, *chart.Config, chartutil.ReleaseOptions, *chartutil.Capabilities) (chartutil.Values, error) {
	return func(chrt *chart.Chart, chrtVals *chart.Config, options chartutil.ReleaseOptions, caps *chartutil.Capabilities) (chartutil.Values, error) {
		assert.Equal(r.t, expChrt, *chrt)
		assert.Equal(r.t, expResp.Release.Config, chrtVals)
		assert.Equal(r.t, chartutil.ReleaseOptions{
			Name:      expResp.Release.Name,
			Time:      expResp.Release.Info.LastDeployed,
			Namespace: expResp.Release.Namespace,
			Revision:  int(expResp.Release.Version),
			IsInstall: true,
		}, options)
		assert.Equal(r.t, &chartutil.Capabilities{}, caps)
		return fixChartutilValues(), nil
	}
}

func (r toRenderValuesFake) WithForcedError(err error) func(*chart.Chart, *chart.Config, chartutil.ReleaseOptions, *chartutil.Capabilities) (chartutil.Values, error) {
	return func(chrt *chart.Chart, chrtVals *chart.Config, options chartutil.ReleaseOptions, caps *chartutil.Capabilities) (chartutil.Values, error) {
		return nil, err
	}
}

func fixChartutilValues() chartutil.Values {
	return chartutil.Values{"fix_val_key": "fix_val"}
}

func fixChart() chart.Chart {
	return chart.Chart{
		Metadata: &chart.Metadata{
			Name: "test-chart",
		},
	}
}
func fixInstallReleaseResponse(ch chart.Chart) *services.InstallReleaseResponse {
	return &services.InstallReleaseResponse{
		Release: &hapi_release5.Release{
			Info: &hapi_release5.Info{
				LastDeployed: &google_protobuf.Timestamp{
					Seconds: 123123123,
					Nanos:   1,
				},
			},
			Config: &chart.Config{
				Raw: "raw-config",
			},
			Name:      "test-release",
			Namespace: "test-ns",
			Version:   int32(123),
			Chart:     &ch,
		},
	}
}
