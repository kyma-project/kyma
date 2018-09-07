package overrides

import (
	"path"

	"github.com/kyma-project/kyma/components/installer/pkg/apis/installer/v1alpha1"
	"github.com/kyma-project/kyma/components/installer/pkg/config"
	"github.com/kyma-project/kyma/components/installer/pkg/errors"
	"github.com/kyma-project/kyma/components/installer/pkg/kymasources"
)

// LegacyProvider .
type LegacyProvider interface {
	GetForRelease(component v1alpha1.KymaComponent) (string, error)
}

type legacyProvider struct {
	overrideData     OverrideData
	installationData *config.InstallationData
	kymaSources      kymasources.KymaPackage
	errorHandlers    errors.ErrorHandlersInterface
}

func (p legacyProvider) GetForRelease(component v1alpha1.KymaComponent) (string, error) {
	chartDir := path.Join(p.kymaSources.GetChartsDirPath(), component.Name)

	allOverrides := Map{}

	MergeMaps(allOverrides, p.overrideData.Common())
	MergeMaps(allOverrides, p.overrideData.ForComponent(component.GetReleaseName()))

	var overridesFunc func(*config.InstallationData, string, Map) (string, error)

	switch component.GetReleaseName() {
	case "cluster-essentials":
		overridesFunc = p.getClusterEssentialsOverrides
		break
	case "istio":
		overridesFunc = p.getIstioOverrides
		break
	case "prometheus-operator":
		overridesFunc = p.getPrometheusOverrides
		break
	case "dex":
		overridesFunc = p.getDexOverrides
		break
	case "core":
		overridesFunc = p.getCoreOverrides
		break
	case "hmc-default":
		overridesFunc = p.getHmcOverrides
		break
	case "ec-default":
		overridesFunc = p.getEcOverrides
		break
	}

	if overridesFunc != nil {
		overridesStr, err := overridesFunc(p.installationData, chartDir, allOverrides)
		if err != nil {
			return "", err
		}

		return overridesStr, nil
	}

	return "", nil
}

func (p legacyProvider) getCoreOverrides(installationData *config.InstallationData, chartDir string, overrides Map) (string, error) {
	allOverrides := Map{}

	MergeMaps(allOverrides, overrides)

	globalOverrides, err := GetGlobalOverrides(installationData, allOverrides)
	p.errorHandlers.LogError("Couldn't get global overrides: ", err)
	MergeMaps(allOverrides, globalOverrides)

	coreOverrides, err := GetCoreOverrides(installationData, allOverrides)
	p.errorHandlers.LogError("Couldn't get Kyma core overrides: ", err)

	MergeMaps(allOverrides, coreOverrides)

	staticOverrides := getStaticFileOverrides(installationData, chartDir)
	if staticOverrides.HasOverrides() == true {
		fileOverrides, err := staticOverrides.GetOverrides()
		p.errorHandlers.LogError("Couldn't get additional overrides: ", err)
		MergeMaps(allOverrides, fileOverrides)
	}

	return ToYaml(allOverrides)
}

func (p legacyProvider) getClusterEssentialsOverrides(installationData *config.InstallationData, chartDir string, overrides Map) (string, error) {
	allOverrides := Map{}

	MergeMaps(allOverrides, overrides)

	globalOverrides, err := GetGlobalOverrides(installationData, allOverrides)
	p.errorHandlers.LogError("Couldn't get global overrides: ", err)
	MergeMaps(allOverrides, globalOverrides)

	staticOverrides := getStaticFileOverrides(installationData, chartDir)
	if staticOverrides.HasOverrides() == true {
		fileOverrides, err := staticOverrides.GetOverrides()
		p.errorHandlers.LogError("Couldn't get additional overrides: ", err)
		MergeMaps(allOverrides, fileOverrides)
	}

	return ToYaml(allOverrides)
}

func (p legacyProvider) getIstioOverrides(installationData *config.InstallationData, chartDir string, overrides Map) (string, error) {
	allOverrides := Map{}

	MergeMaps(allOverrides, overrides)

	globalOverrides, err := GetGlobalOverrides(installationData, allOverrides)
	p.errorHandlers.LogError("Couldn't get global overrides: ", err)
	MergeMaps(allOverrides, globalOverrides)

	istioOverrides, err := GetIstioOverrides(installationData, allOverrides)
	p.errorHandlers.LogError("Couldn't get Istio overrides: ", err)
	MergeMaps(allOverrides, istioOverrides)

	staticOverrides := getStaticFileOverrides(installationData, chartDir)
	if staticOverrides.HasOverrides() == true {
		fileOverrides, err := staticOverrides.GetOverrides()
		p.errorHandlers.LogError("Couldn't get additional overrides: ", err)
		MergeMaps(allOverrides, fileOverrides)
	}

	return ToYaml(allOverrides)
}

func (p legacyProvider) getPrometheusOverrides(installationData *config.InstallationData, chartDir string, overrides Map) (string, error) {
	//TODO: this does not get globalOverrides... Is that a problem if global will carry all external ones (overrides.yaml + from configMaps/secrets?)
	allOverrides := Map{}

	MergeMaps(allOverrides, overrides)

	staticOverrides := getStaticFileOverrides(installationData, chartDir)
	if staticOverrides.HasOverrides() == true {
		fileOverrides, err := staticOverrides.GetOverrides()
		p.errorHandlers.LogError("Couldn't get additional overrides: ", err)
		MergeMaps(allOverrides, fileOverrides)
	}

	return ToYaml(allOverrides)
}

func (p legacyProvider) getDexOverrides(installationData *config.InstallationData, chartDir string, overrides Map) (string, error) {

	allOverrides := Map{}

	MergeMaps(allOverrides, overrides)

	globalOverrides, err := GetGlobalOverrides(installationData, allOverrides)
	p.errorHandlers.LogError("Couldn't get global overrides: ", err)
	MergeMaps(allOverrides, globalOverrides)

	staticOverrides := getStaticFileOverrides(installationData, chartDir)
	if staticOverrides.HasOverrides() == true {
		fileOverrides, err := staticOverrides.GetOverrides()
		p.errorHandlers.LogError("Couldn't get additional overrides: ", err)
		MergeMaps(allOverrides, fileOverrides)
	}

	return ToYaml(allOverrides)
}

func (p legacyProvider) getHmcOverrides(installationData *config.InstallationData, chartDir string, overrides Map) (string, error) {
	allOverrides := Map{}

	MergeMaps(allOverrides, overrides)

	globalOverrides, err := GetGlobalOverrides(installationData, allOverrides)
	p.errorHandlers.LogError("Couldn't get global overrides: ", err)
	MergeMaps(allOverrides, globalOverrides)

	staticOverrides := getStaticFileOverrides(installationData, chartDir)
	if staticOverrides.HasOverrides() == true {
		fileOverrides, err := staticOverrides.GetOverrides()
		p.errorHandlers.LogError("Couldn't get additional overrides: ", err)
		MergeMaps(allOverrides, fileOverrides)
	}

	return ToYaml(allOverrides)
}

func (p legacyProvider) getEcOverrides(installationData *config.InstallationData, chartDir string, overrides Map) (string, error) {
	allOverrides := Map{}

	MergeMaps(allOverrides, overrides)

	globalOverrides, err := GetGlobalOverrides(installationData, allOverrides)
	p.errorHandlers.LogError("Couldn't get global overrides: ", err)
	MergeMaps(allOverrides, globalOverrides)

	staticOverrides := getStaticFileOverrides(installationData, chartDir)
	if staticOverrides.HasOverrides() == true {
		fileOverrides, err := staticOverrides.GetOverrides()
		p.errorHandlers.LogError("Couldn't get additional overrides: ", err)
		MergeMaps(allOverrides, fileOverrides)
	}

	return ToYaml(allOverrides)
}

// NewLegacyProvider .
func NewLegacyProvider(overrideData OverrideData, installationData *config.InstallationData, kymaSources kymasources.KymaPackage, errorHandlers errors.ErrorHandlersInterface) LegacyProvider {
	return legacyProvider{
		overrideData:     overrideData,
		installationData: installationData,
		kymaSources:      kymaSources,
		errorHandlers:    errorHandlers,
	}
}

func getStaticFileOverrides(installationData *config.InstallationData, chartDir string) StaticFile {
	if installationData.IsLocalInstallation {
		return NewLocalStaticFile()
	}

	return NewClusterStaticFile(chartDir)
}
