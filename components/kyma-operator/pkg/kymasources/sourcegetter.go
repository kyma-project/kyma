package kymasources

import (
	"context"
	"log"
	"os"
	"path"

	"github.com/hashicorp/go-getter"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
)

const (
	//Subdirectory used to store sources fetched by user-defined remote sources feature
	componentsRemoteSrcDir = "components-remote-src"
)

// SourceGetter defines contract for fetching component sources
type SourceGetter interface {
	// SrcDirFor returns a local filesystem directory path to the component sources.
	// If the component is configured with external `Source.URL`, it's sources are first downloaded to a local directory.
	// Otherwise component sources bundled with kyma-operator Docker image are used.
	SrcDirFor(component v1alpha1.KymaComponent) (string, error)
}

type SourceGetterCreator interface {
	//Required only to support legacy mechanism of downloading Kyma sources using URL/KymaVersion attributes of InstallationData.
	NewGetterFor(url, version string) SourceGetter
}

//Used to create instances of componentSrcGetter
//componentSrcGetter still depends on Installation CR parameters, because of legacy mechanism URL/KymaVersion.
//Once the legacy mechanism is removed, componentSrcGetter does not depend on Installation CR instance anymore and this type can also be removed.
type sourceGetterCreator struct {
	defaultSourcesHandler *defaultSources
	fsWrapper             FilesystemWrapper
	rootDir               string
}

func NewSourceGetterCreator(kymaPackages KymaPackages, fsWrapper FilesystemWrapper, rootDir string) SourceGetterCreator {
	defaultSourcesHandler := newDefaultSources(kymaPackages)

	return &sourceGetterCreator{
		defaultSourcesHandler,
		fsWrapper,
		rootDir,
	}
}

//NewGetterFor returns a SourceGetter configured with "url" and "kymaVersion" taken from Installation CR instance.
func (sgc *sourceGetterCreator) NewGetterFor(url string, kymaVersion string) SourceGetter {

	//In case ALL components have user-defined sources, we don't have to fallback to "defaults" at all. That's why invocation of defaultSourcesHandler is wrapped in a funtion - this allows for lazy usage.
	defPkgFn := func() (KymaPackage, error) {
		res, err := sgc.defaultSourcesHandler.ensureDefaultSources(url, kymaVersion)
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
		log.Printf("Component %s configured with remote sources", component.Name)
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

	componentRemoteSrcDir := path.Join(csr.rootDir, componentsRemoteSrcDir, component.Name)
	if !csr.fsWrapper.Exists(componentRemoteSrcDir) {
		log.Printf("Remote sources for component %s do not exist", component.Name)
		return csr.getSourcesFor(component, componentRemoteSrcDir)
	}

	log.Printf("Remote sources for component %s already cached, reusing.", component.Name)
	return componentRemoteSrcDir, nil
}

func (csr *componentSrcGetter) getSourcesFor(component v1alpha1.KymaComponent, destDir string) (string, error) {

	log.Printf("Fetching sources for component %s from remote location", component.Name)
	if !csr.fsWrapper.Exists(destDir) {
		if err := csr.fsWrapper.CreateDir(destDir); err != nil {
			return "", err
		}
	}

	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	src := component.Source.URL
	dst := destDir
	mode := getter.ClientModeDir
	opts := []getter.ClientOption{}

	var client = &getter.Client{
		Ctx:     ctx,
		Src:     src,
		Dst:     dst,
		Pwd:     pwd,
		Mode:    mode,
		Options: opts,
	}

	err = client.Get()
	if err != nil {
		return "", err
	}

	log.Printf("Sources for component %s fetched from remote location", component.Name)
	return destDir, nil
}
