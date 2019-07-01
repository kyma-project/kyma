package main

import (
	"flag"
	"fmt"

	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	bu "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	apiErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	removeOwnerReference  = "removeOwnerReference"
	addOwnerReference     = "addOwnerReference"
	actionFlag            = "action"
	apiServerNameFlag     = "apiServerName"
	sbuControllerNameFlag = "sbuControllerName"
	namespaceFlag         = "namespace"
)

type config struct {
	action            string
	apiServerName     string
	sbuControllerName string
	namespace         string
}

type sbuManager struct {
	client   kubernetes.Interface
	buClient bu.Interface
	scClient sc.Interface
}

func main() {
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("failed during get config: %s", err)
	}
	client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		log.Fatalf("failed during get kubernetes client: %s", err)
	}
	buClient, err := bu.NewForConfig(kubeConfig)
	if err != nil {
		log.Fatalf("failed during get service binding usage client: %s", err)
	}
	scClient, err := sc.NewForConfig(kubeConfig)
	if err != nil {
		log.Fatalf("failed during get service catalog client: %s", err)
	}
	cfg, err := parseFlag()
	if err != nil {
		log.Fatalf("failed during parsing flags: %s", err)
	}
	manager := sbuManager{client: client, buClient: buClient, scClient: scClient}

	switch cfg.action {
	case removeOwnerReference:
		migrationRequired, err := manager.isMigrationRequired(cfg.apiServerName, cfg.namespace)
		if err != nil {
			log.Fatalf("failed during assessing run process: %s", err)
		}
		if !migrationRequired {
			log.Info("Process is not required.")
			return
		}
		log.Info("Scale down ServiceBindingUsage controller")
		if err = manager.scaleDownSBUController(cfg.sbuControllerName, cfg.namespace); err != nil {
			log.Fatalf("while scaling down SBU controller: %s", err)
		}
		log.Info("Start removing owner reference from ServiceBindingUsage")
		if err := manager.prepareServiceBindingUsages(); err != nil {
			log.Fatalf("while prepare service binding usage to upgrade: %s", err)
		}
		log.Info("Removing owner reference from ServiceBindingUsage is done")
	case addOwnerReference:
		log.Info("Scale up ServiceBindingUsage controller")
		if err = manager.scaleUpSBUController(cfg.sbuControllerName, cfg.namespace); err != nil {
			log.Fatalf("while scaling up SBU controller: %s", err)
		}
		log.Info("Start restoring owner reference to ServiceBindingUsage")
		if err := manager.restoreServiceBindingUsages(); err != nil {
			log.Fatalf("while restore service binding usage after upgrade: %s", err)
		}
		log.Info("Restoring owner reference to ServiceBindingUsage is done")
	default:
		log.Fatalf("The command is not support. Available commands: %s, %s", removeOwnerReference, addOwnerReference)
	}
}

func parseFlag() (*config, error) {
	var cfg config

	flag.StringVar(&cfg.action, actionFlag, "", "Name of process to run")
	flag.StringVar(&cfg.apiServerName, apiServerNameFlag, "", "The name of ApiServer deployment")
	flag.StringVar(&cfg.sbuControllerName, sbuControllerNameFlag, "", "The name of SBU controller deployment")
	flag.StringVar(&cfg.namespace, namespaceFlag, "", "Release namespace")
	flag.Parse()

	switch cfg.action {
	case removeOwnerReference:
		if cfg.apiServerName == "" {
			return nil, fmt.Errorf("Flag %q is required", apiServerNameFlag)
		}
		if cfg.sbuControllerName == "" {
			return nil, fmt.Errorf("Flag %q is required", sbuControllerNameFlag)
		}
		if cfg.namespace == "" {
			return nil, fmt.Errorf("Flag %q is required", namespaceFlag)
		}
	case addOwnerReference:
		if cfg.sbuControllerName == "" {
			return nil, fmt.Errorf("Flag %q is required", sbuControllerNameFlag)
		}
		if cfg.namespace == "" {
			return nil, fmt.Errorf("Flag %q is required", namespaceFlag)
		}
	}

	return &cfg, nil
}

func (sm *sbuManager) isMigrationRequired(name, namespace string) (bool, error) {
	_, err := sm.client.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
	if apiErr.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, errors.Wrap(err, "while fetching ApiServer deployment")
	}

	return true, nil
}

func (sm *sbuManager) scaleDownSBUController(name, namespace string) error {
	replicas := int32(0)

	return sm.scaleSBUController(name, namespace, replicas)
}

func (sm *sbuManager) scaleUpSBUController(name, namespace string) error {
	replicas := int32(1)

	return sm.scaleSBUController(name, namespace, replicas)
}

func (sm *sbuManager) scaleSBUController(name, namespace string, replicas int32) error {
	sbuController, err := sm.client.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "while fetching SBU Controller deployment")
	}

	deploymentCopy := sbuController.DeepCopy()
	deploymentCopy.Spec.Replicas = &replicas

	_, err = sm.client.AppsV1().Deployments(namespace).Update(deploymentCopy)
	if err != nil {
		return errors.Wrap(err, "while updating SBU Controller deployment")
	}

	return nil
}

func (sm *sbuManager) prepareServiceBindingUsages() error {
	list, err := sm.buClient.ServicecatalogV1alpha1().ServiceBindingUsages(metav1.NamespaceAll).List(metav1.ListOptions{})
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

	_, err := sm.buClient.ServicecatalogV1alpha1().ServiceBindingUsages(toUpdate.Namespace).Update(toUpdate)
	if err != nil {
		return errors.Wrapf(err, "while updating %q ServiceBindingUsage", toUpdate.Name)
	}

	return nil
}

func (sm *sbuManager) restoreServiceBindingUsages() error {
	listSBU, err := sm.buClient.ServicecatalogV1alpha1().ServiceBindingUsages(metav1.NamespaceAll).List(metav1.ListOptions{})
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
	list, err := sm.scClient.ServicecatalogV1beta1().ServiceBindings(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "while fetching ServiceBinding list")
	}

	for _, sb := range list.Items {
		if name == sb.Name && namespace == sb.Namespace {
			return &sb, nil
		}
	}

	return nil, fmt.Errorf("there is no %s/%s ServiceBinding", name, namespace)
}

func (sm *sbuManager) updateOwnerReference(sbu v1alpha1.ServiceBindingUsage, sb *v1beta1.ServiceBinding) error {
	toUpdate := sbu.DeepCopy()
	toUpdate.OwnerReferences = append(toUpdate.OwnerReferences, metav1.OwnerReference{
		APIVersion: "servicecatalog.k8s.io/v1beta1",
		Kind:       "ServiceBinding",
		Name:       sb.Name,
		UID:        sb.UID,
	})

	_, err := sm.buClient.ServicecatalogV1alpha1().ServiceBindingUsages(toUpdate.Namespace).Update(toUpdate)
	if err != nil {
		return errors.Wrapf(err, "while updating %q ServiceBindingUsage", toUpdate.Name)
	}

	return nil
}
