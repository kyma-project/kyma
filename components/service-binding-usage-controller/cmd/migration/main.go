package main

import (
	"fmt"
	"os"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	bu "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

const (
	removeOwnerReference = "removeOwnerReference"
	addOwnerReference    = "addOwnerReference"
)

type sbuManager struct {
	client   bu.Interface
	scClient sc.Interface
	sbList   *v1beta1.ServiceBindingList
}

func main() {
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("failed during get config: %s", err)
	}
	client, err := bu.NewForConfig(kubeConfig)
	if err != nil {
		log.Fatalf("failed during get service binding usage client: %s", err)
	}
	scClient, err := sc.NewForConfig(kubeConfig)
	if err != nil {
		log.Fatalf("failed during get service catalog client: %s", err)
	}

	args := os.Args
	manager := sbuManager{client: client, scClient: scClient}
	switch args[1] {
	case removeOwnerReference:
		log.Info("Start removing owner reference from ServiceBindingUsage")
		if err := manager.prepareServiceBindingUsages(); err != nil {
			log.Fatalf("while prepare service binding usage to upgrade: %s", err)
		}
		log.Info("Removing owner reference from ServiceBindingUsage is done")
	case addOwnerReference:
		log.Info("Start restoring owner reference to ServiceBindingUsage")
		if err := manager.restoreServiceBindingUsages(); err != nil {
			log.Fatalf("while restore service binding usage after upgrade: %s", err)
		}
		log.Info("Restoring owner reference to ServiceBindingUsage is done")
	default:
		log.Fatalf("The command is not support. Available commands: %s, %s", removeOwnerReference, addOwnerReference)
	}
}

func (sm *sbuManager) prepareServiceBindingUsages() error {
	list, err := sm.client.ServicecatalogV1alpha1().ServiceBindingUsages(v1.NamespaceAll).List(v1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "while fetching ServiceBindingUsage list")
	}

	for _, sbu := range list.Items {
		err = sm.removeOwnerReference(&sbu)
		if err != nil {
			return errors.Wrapf(err, "while removing owner reference from %q ServiceBindingUsage", sbu.Name)
		}
	}

	return nil
}

func (sm *sbuManager) removeOwnerReference(sbu *v1alpha1.ServiceBindingUsage) error {
	toUpdate := sbu.DeepCopy()
	toUpdate.ObjectMeta.OwnerReferences = nil

	_, err := sm.client.ServicecatalogV1alpha1().ServiceBindingUsages(toUpdate.Namespace).Update(toUpdate)
	if err != nil {
		return errors.Wrapf(err, "while updating %q ServiceBindingUsage", toUpdate.Name)
	}

	return nil
}

func (sm *sbuManager) restoreServiceBindingUsages() error {
	listSBU, err := sm.client.ServicecatalogV1alpha1().ServiceBindingUsages(v1.NamespaceAll).List(v1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "")
	}

	for _, sbu := range listSBU.Items {
		if sm.hasOwnerReference(sbu) {
			continue
		}
		sb, err := sm.findServiceBinding(sbu.Spec.ServiceBindingRef.Name, sbu.Namespace)
		if err != nil {
			return errors.Wrapf(err, "while finding ServiceBinding %q", sbu.Spec.ServiceBindingRef.Name)
		}
		err = sm.updateOwnerReference(sbu, sb)
		if err != nil {
			return nil
		}
	}

	return nil
}

func (sm *sbuManager) hasOwnerReference(sbu v1alpha1.ServiceBindingUsage) bool {
	for _, or := range sbu.OwnerReferences {
		if or.Name == sbu.Spec.ServiceBindingRef.Name {
			return true
		}
	}

	return false
}

func (sm *sbuManager) findServiceBinding(name, namespace string) (*v1beta1.ServiceBinding, error) {
	if sm.sbList == nil {
		list, err := sm.scClient.ServicecatalogV1beta1().ServiceBindings(v1.NamespaceAll).List(v1.ListOptions{})
		if err != nil {
			return nil, errors.Wrap(err, "while fetching ServiceBinding list")
		}
		sm.sbList = list
	}

	for _, sb := range sm.sbList.Items {
		if name == sb.Name && namespace == sb.Namespace {
			return &sb, nil
		}
	}

	return nil, fmt.Errorf("there is no %s/%s ServiceBinding", name, namespace)
}

func (sm *sbuManager) updateOwnerReference(sbu v1alpha1.ServiceBindingUsage, sb *v1beta1.ServiceBinding) error {
	toUpdate := sbu.DeepCopy()
	toUpdate.OwnerReferences = append(toUpdate.OwnerReferences, metaV1.OwnerReference{
		APIVersion: "servicecatalog.k8s.io/v1beta1",
		Kind:       "ServiceBinding",
		Name:       sb.Name,
		UID:        sb.UID,
	})

	_, err := sm.client.ServicecatalogV1alpha1().ServiceBindingUsages(toUpdate.Namespace).Update(toUpdate)
	if err != nil {
		return errors.Wrapf(err, "while updating %q ServiceBindingUsage", toUpdate.Name)
	}

	return nil
}
