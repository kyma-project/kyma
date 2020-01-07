package kymasources

import (
	"errors"
	"log"

	internalerrors "github.com/kyma-project/kyma/components/kyma-operator/pkg/errors"
)

//defaultSources handles default Kyma sources (bundled or downloaded using legacy URL/version mechanism)
type defaultSources struct {
	kymaPackages  KymaPackages
	errorHandlers internalerrors.ErrorHandlersInterface
}

func newDefaultSources(kymaPackages KymaPackages) *defaultSources {
	return &defaultSources{
		kymaPackages:  kymaPackages,
		errorHandlers: &internalerrors.ErrorHandlers{},
	}
}

//ensureDefaultSources ensures Kyma sources are bundled with the kyma-operator or downloads them
func (ds defaultSources) ensureDefaultSources(url, kymaVersion string) (KymaPackage, error) {

	const operation string = "Get Kyma Sources"
	log.Println(operation)

	if ds.kymaPackages.HasBundledSources() {
		log.Println("Kyma sources available locally.")
		log.Println(operation + "...DONE")

		return ds.kymaPackages.GetBundledPackage()
	}

	//TODO: Deprecated. Sources can be specified per component with `Source.URL` property.
	log.Println("Kyma sources not available. Downloading...")

	if kymaVersion == "" {
		validationErr := errors.New("set version for Kyma package")
		//ds.errorHandlers.LogError("Validation error: ", validationErr)
		//_ = ds.statusManager.Error("Kyma Operator", operation, validationErr)
		return nil, validationErr
	}

	if url == "" {
		validationErr := errors.New("set url to Kyma package")
		//ds.errorHandlers.LogError("Validation error: ", validationErr)
		//_ = ds.statusManager.Error("Kyma Operator", operation, validationErr)
		return nil, validationErr
	}

	log.Println("Downloading Kyma, Version: " + kymaVersion + " URL: " + url)

	err := ds.kymaPackages.FetchPackage(url, kymaVersion)
	if ds.errorHandlers.CheckError("Fetch Kyma package error: ", err) {
		//_ = ds.statusManager.Error("Kyma Operator", operation, err)
		return nil, err
	}

	kymaPackage, err := ds.kymaPackages.GetPackage(kymaVersion)
	if ds.errorHandlers.CheckError("Get Kyma package error: ", err) {
		//_ = ds.statusManager.Error("Kyma Operator", operation, err)
		return nil, err
	}

	log.Println(operation + "...DONE")

	return kymaPackage, nil
}
