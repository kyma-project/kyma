package kymasources

import (
	"context"
	"log"
	"os"
	"path"
	"time"

	retry "github.com/avast/retry-go"
	"github.com/hashicorp/go-getter"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
)

const (
	//Subdirectory used to store sources fetched by user-defined remote sources feature
	componentsRemoteSrcDir = "components-remote-src"
)

//Config for legacy mechanism for fetching remote Kyma sources from a remote location.
type LegacyKymaSourceConfig struct {
	KymaURL     string
	KymaVersion string
}

// SourceGetter defines contract for resolving component sources
type SourceGetter interface {
	// SrcDirFor returns a local filesystem directory path to the component sources.
	// If the component is configured with external `Source.URL`, it's sources are first downloaded to a local directory.
	// Otherwise component sources bundled with kyma-operator Docker image are used.
	SrcDirFor(component v1alpha1.KymaComponent) (string, error)
}

//SourceGetterCreator is used to create a SourceGetter instance.
//In order to support legacy mechanism of fetching Kyma sources for entire installation from remote location, it depends on KymaUrl/Kyma version parameters from Installation CR.
//TODO: Once URL/KymaVersion are removed from Installation CR, SourceGetter no longer depends on the Installation CR instance and this interface can be removed.
type SourceGetterCreator interface {
	NewGetterFor(legacyKymaSourceConfig LegacyKymaSourceConfig) SourceGetter
}

//Used to create instances of componentSrcGetter
//componentSrcGetter still depends on Installation CR parameters, because of legacy mechanism URL/KymaVersion.
//Once the legacy mechanism is removed, componentSrcGetter does not depend on Installation CR instance anymore and this type can also be removed.
type sourceGetterCreator struct {
	defaultSourcesHandler *defaultSources
	fsWrapper             FilesystemWrapper
	rootDir               string
}

//NewSourceGetterCreator returns an instance of SourceGetterCreator
func NewSourceGetterCreator(kymaPackages KymaPackages, fsWrapper FilesystemWrapper, rootDir string) SourceGetterCreator {
	defaultSourcesHandler := newDefaultSources(kymaPackages)

	return &sourceGetterCreator{
		defaultSourcesHandler,
		fsWrapper,
		rootDir,
	}
}

//NewGetterFor returns a SourceGetter configured with "url" and "kymaVersion" taken from Installation CR instance.
func (sgc *sourceGetterCreator) NewGetterFor(legacyKymaSourceConfig LegacyKymaSourceConfig) SourceGetter {

	//In case ALL components have user-defined sources, we don't have to fallback to "defaults" at all. That's why defaultSourcesHandler is invoked lazily with the help of this function.
	defPkgFn := func() (KymaPackage, error) {
		res, err := sgc.defaultSourcesHandler.ensureDefaultSources(legacyKymaSourceConfig.KymaURL, legacyKymaSourceConfig.KymaVersion)
		if err != nil {
			return nil, err
		}
		return res, nil
	}

	return &componentSrcGetter{
		fsWrapper:        sgc.fsWrapper,
		rootDir:          sgc.rootDir,
		defaultPackageFn: defPkgFn,
	}
}

//componentSrcGetter is used to get component sources necessary for the installation
//Implements SourceGetter
type componentSrcGetter struct {
	fsWrapper        FilesystemWrapper
	rootDir          string
	defaultPackageFn func() (KymaPackage, error) //Used lazily - only when necessary
}

//SourceDirFor returns a directory path with charts for given component
func (csr *componentSrcGetter) SrcDirFor(component v1alpha1.KymaComponent) (string, error) {

	if component.Source != nil && len(component.Source.URL) > 0 {
		//Handle user-defined component sources
		log.Printf("Component \"%s\" configured with remote sources", component.Name)
		return csr.ensureRemoteSourcesFor(component)
	}

	//Fallback to defaults. These are the sources bundled with the Kyma-operator or downloaded from an external location
	defaultKymaSources, err := csr.defaultPackageFn()
	if err != nil {
		return "", err
	}
	var defaultSourcesRoot = defaultKymaSources.GetChartsDirPath()
	componentDir := path.Join(defaultSourcesRoot, component.Name)
	return componentDir, nil
}

func (csr *componentSrcGetter) ensureRemoteSourcesFor(component v1alpha1.KymaComponent) (string, error) {
	componentSrcLocalCopyDir := path.Join(csr.rootDir, componentsRemoteSrcDir, component.Name)
	componentChartFile := path.Join(componentSrcLocalCopyDir, "Chart.yaml")
	if !csr.fsWrapper.Exists(componentChartFile) {
		log.Printf("Remote sources for component \"%s\" do not exist", component.Name)
		return csr.getSourcesFor(component, componentSrcLocalCopyDir)
	}

	log.Printf("Remote sources for component \"%s\" already cached, reusing.", component.Name)
	return componentSrcLocalCopyDir, nil
}

func (csr *componentSrcGetter) getSourcesFor(component v1alpha1.KymaComponent, destDir string) (string, error) {

	log.Printf("Fetching sources for component \"%s\" from remote location", component.Name)
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	var client = &getter.Client{
		Ctx:  context.Background(),
		Src:  component.Source.URL,
		Dst:  destDir,
		Pwd:  pwd,
		Mode: getter.ClientModeDir,
	}

	const maxAttempts = 3
	const delay = time.Second * 10

	err = retry.Do(
		func() error {
			return client.Get()
		},
		retry.Attempts(maxAttempts),
		retry.Delay(delay),
		retry.DelayType(retry.FixedDelay),
		retry.OnRetry(func(retryNo uint, err error) {
			log.Printf("Retry fetching component sources: [%d / %d], error: %s", retryNo+1, maxAttempts, err)
		}),
	)

	if err != nil {
		return "", err
	}

	log.Printf("Using remote sources for installation of component \"%s\"", component.Name)
	return destDir, nil
}
