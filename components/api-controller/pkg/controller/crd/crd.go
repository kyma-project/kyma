package crd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	apiExtensionsApi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiExtensions "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Registrar struct {
	apiExtensionsInterface apiExtensions.Interface
}

func NewRegistrar(apiExtensionsInterface apiExtensions.Interface) *Registrar {
	return &Registrar{apiExtensionsInterface: apiExtensionsInterface}
}

func (r *Registrar) Register(crd *apiExtensionsApi.CustomResourceDefinition) {

	crdLabel := fmt.Sprintf("%s/%s", crd.ObjectMeta.Name, crd.Spec.Version)

	log.Infof("Creating CRD '%s'...", crdLabel)
	_, err := r.CustomResourceDefinitions().Create(crd)

	if err != nil {

		if apiErrors.IsAlreadyExists(err) {

			log.Infof("CRD '%s' already exists - updating...", crdLabel)

			existingCrd, getErr := r.CustomResourceDefinitions().Get(crd.Name, k8sMeta.GetOptions{})
			if getErr != nil {
				log.Errorf("unable to get existing CRD '%s'. Error: %v", crdLabel, getErr)
				return
			}

			crd.ResourceVersion = existingCrd.ResourceVersion

			_, updateErr := r.CustomResourceDefinitions().Update(crd)
			if updateErr != nil {
				log.Errorf("unable to update existing CRD '%s'. Error: %v", crdLabel, updateErr)
				return
			}
		}

		log.Errorf("unable to create CRD '%s'. Error: %v", crdLabel, err)
		return
	}

	log.Infof("CRD '%s' successfully created.", crdLabel)
}
func (r *Registrar) CustomResourceDefinitions() v1beta1.CustomResourceDefinitionInterface {
	return r.apiExtensionsInterface.ApiextensionsV1beta1().CustomResourceDefinitions()
}
