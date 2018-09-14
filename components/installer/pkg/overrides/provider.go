package overrides

import (
	"path"

	"github.com/kyma-project/kyma/components/installer/pkg/apis/installer/v1alpha1"
	"github.com/kyma-project/kyma/components/installer/pkg/errors"
	"github.com/kyma-project/kyma/components/installer/pkg/kymasources"
)

// LegacyProvider .
type LegacyProvider interface {
	GetForRelease(component v1alpha1.KymaComponent) (string, error)
}

type legacyProvider struct {
	overrideData  OverrideData
	kymaSources   kymasources.KymaPackage
	errorHandlers errors.ErrorHandlersInterface
}

func (p legacyProvider) GetForRelease(component v1alpha1.KymaComponent) (string, error) {
	chartDir := path.Join(p.kymaSources.GetChartsDirPath(), component.Name)

	allOverrides := Map{}

	MergeMaps(allOverrides, p.overrideData.Common())
	MergeMaps(allOverrides, p.overrideData.ForComponent(component.GetReleaseName()))

	overridesStr, err := p.getOverrides(chartDir, allOverrides)

	if err != nil {
		return "", err
	}

	return overridesStr, nil
}

func (p legacyProvider) getOverrides(chartDir string, overrides Map) (string, error) {
	allOverrides := Map{}

	MergeMaps(allOverrides, overrides)

	staticOverrides := p.getStaticFileOverrides(overrides, chartDir)
	if staticOverrides.HasOverrides() == true {
		fileOverrides, err := staticOverrides.GetOverrides()
		p.errorHandlers.LogError("Couldn't get additional overrides: ", err)
		MergeMaps(allOverrides, fileOverrides)
	}

	return ToYaml(allOverrides)
}

// NewLegacyProvider .
func NewLegacyProvider(overrideData OverrideData, kymaSources kymasources.KymaPackage, errorHandlers errors.ErrorHandlersInterface) LegacyProvider {
	return legacyProvider{
		overrideData:  overrideData,
		kymaSources:   kymaSources,
		errorHandlers: errorHandlers,
	}
}

func (p legacyProvider) getStaticFileOverrides(overrides Map, chartDir string) StaticFile {
	isLocalEnv := FindOverrideValue(overrides, "global.isLocalEnv")

	isLocalInst, isBool := isLocalEnv.(bool)

	if isBool && isLocalInst {
		return NewLocalStaticFile()
	}

	return NewClusterStaticFile(chartDir)
}
