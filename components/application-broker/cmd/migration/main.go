package main

import (
	"flag"
	"fmt"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	sc "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	eaCli "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	apiErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	removeOwnerReference     = "removeOwnerReference"
	addOwnerReference        = "addOwnerReference"
	actionFlag               = "action"
	apiServerNameFlag        = "apiServerName"
	namespaceFlag            = "namespace"
	configMapNameForOwnerRef = "storage-for-owner-ref-from-event-activation"
)

type config struct {
	action        string
	apiServerName string
	namespace     string
}

type migrationManager struct {
	workingNS string

	k8sClient kubernetes.Interface
	eaClient  eaCli.ApplicationconnectorV1alpha1Interface
	scClient  sc.Interface
}

func main() {
	kubeConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("while creating k8s config: %s", err)
	}
	client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		log.Fatalf("cannot create kubernetes client: %s", err)
	}
	eaClient, err := eaCli.NewForConfig(kubeConfig)
	if err != nil {
		log.Fatalf("cannot create event activation client: %s", err)
	}
	scClient, err := sc.NewForConfig(kubeConfig)
	if err != nil {
		log.Fatalf("cannot create service catalog client: %s", err)
	}
	cfg, err := parseFlag()
	if err != nil {
		log.Fatalf("while parsing flags: %s", err)
	}
	manager := migrationManager{k8sClient: client, eaClient: eaClient, scClient: scClient, workingNS: cfg.namespace}

	switch cfg.action {
	case removeOwnerReference:
		log.Info("Start removing owner reference from EventActivation")
		if err = manager.ensureBackupConfigMapExists(); err != nil {
			log.Fatalf("while ensuring that backup config map exists: %s", err)
		}
		migrationRequired, err := manager.isMigrationRequired(cfg.apiServerName, cfg.namespace)
		if err != nil {
			log.Fatalf("failed during assessing run process: %s", err)
		}
		if !migrationRequired {
			log.Info("Process is not required.")
			return
		}

		log.Info("Start removing owner reference from EventActivation")
		if err := manager.removeOwnerRefFromEventActivations(); err != nil {
			log.Fatalf("while prepare EventActivation to upgrade: %s", err)
		}
		log.Info("Removing owner reference from EventActivation is done")
	case addOwnerReference:
		log.Info("Start restoring owner reference to EventActivations")
		if err := manager.restoreOwnerRefForEventActivations(); err != nil {
			log.Fatalf("while restoring EventActivation after upgrade: %s", err)
		}
		log.Info("Restoring owner reference to EventActivations is done")
	default:
		log.Fatalf("The command is not support. Available commands: %s, %s", removeOwnerReference, addOwnerReference)
	}
}

func parseFlag() (*config, error) {
	var cfg config

	flag.StringVar(&cfg.action, actionFlag, "", "Name of process to run")
	flag.StringVar(&cfg.apiServerName, apiServerNameFlag, "", "The name of ApiServer deployment")
	flag.StringVar(&cfg.namespace, namespaceFlag, "", "Release namespace")
	flag.Parse()

	switch cfg.action {
	case removeOwnerReference:
		if cfg.apiServerName == "" {
			return nil, fmt.Errorf("Flag %q is required", apiServerNameFlag)
		}
		if cfg.namespace == "" {
			return nil, fmt.Errorf("Flag %q is required", namespaceFlag)
		}
	case addOwnerReference:
		if cfg.namespace == "" {
			return nil, fmt.Errorf("Flag %q is required", namespaceFlag)
		}
	default:
		log.Fatalf("The command is not support. Available commands: %s, %s", removeOwnerReference, addOwnerReference)
	}

	return &cfg, nil
}

func (sm *migrationManager) isMigrationRequired(name, namespace string) (bool, error) {
	_, err := sm.k8sClient.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
	if apiErr.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, errors.Wrap(err, "while fetching ApiServer deployment")
	}

	return true, nil
}

func (sm *migrationManager) removeOwnerRefFromEventActivations() error {
	list, err := sm.eaClient.EventActivations(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "while fetching EventActivations list")
	}

	for _, ea := range list.Items {
		if len(ea.OwnerReferences) == 0 {
			log.Infof("Skipping EventActivation %s/%s because it does not have any owner reference", ea.Namespace, ea.Name)
			continue
		}

		if err := sm.backupOwnerRefFromEventActivation(&ea); err != nil {
			return errors.Wrapf(err, "while storing owner reference from %q EventActivation", ea.Name)
		}

		if err = sm.removeOwnerReference(&ea); err != nil {
			return errors.Wrapf(err, "while removing owner reference from %q EventActivation", ea.Name)
		}
	}

	return nil
}

func (sm *migrationManager) removeOwnerReference(sbu *v1alpha1.EventActivation) error {
	toUpdate := sbu.DeepCopy()

	var refsWithoutSI []metav1.OwnerReference
	for _, oRef := range toUpdate.ObjectMeta.OwnerReferences {
		if oRef.Kind != "ServiceInstance" {
			refsWithoutSI = append(refsWithoutSI, oRef)
		}
	}

	toUpdate.ObjectMeta.OwnerReferences = refsWithoutSI

	_, err := sm.eaClient.EventActivations(toUpdate.Namespace).Update(toUpdate)
	if err != nil {
		return errors.Wrapf(err, "while updating %q EventActivation", toUpdate.Name)
	}

	return nil
}

func (sm *migrationManager) restoreOwnerRefForEventActivations() error {
	listEA, err := sm.eaClient.EventActivations(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "while getting all event activations")
	}
	ownerRefBackup, err := sm.k8sClient.CoreV1().ConfigMaps(sm.workingNS).Get(configMapNameForOwnerRef, metav1.GetOptions{})
	if err != nil {
		return err
	}

	for _, ea := range listEA.Items {
		instanceName, found := ownerRefBackup.Data[eventActivationKey(&ea)]
		if !found {
			continue
		}

		si, err := sm.scClient.ServicecatalogV1beta1().ServiceInstances(ea.Namespace).Get(instanceName, metav1.GetOptions{})
		if err != nil {
			return errors.Wrapf(err, "while fetching ServiceInstance %s/%s", ea.Namespace, instanceName)
		}
		err = sm.updateOwnerReference(ea, si)
		if err != nil {
			return nil
		}
	}

	err = sm.k8sClient.CoreV1().ConfigMaps(sm.workingNS).Delete(configMapNameForOwnerRef, &metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrapf(err, "while deleting config map %q", configMapNameForOwnerRef)
	}

	return nil
}

func (sm *migrationManager) updateOwnerReference(ea v1alpha1.EventActivation, si *v1beta1.ServiceInstance) error {
	toUpdate := ea.DeepCopy()
	toUpdate.OwnerReferences = append(toUpdate.OwnerReferences, metav1.OwnerReference{
		APIVersion: "servicecatalog.k8s.io/v1beta1",
		Kind:       "ServiceInstance",
		Name:       si.Name,
		UID:        si.UID,
	})

	_, err := sm.eaClient.EventActivations(toUpdate.Namespace).Update(toUpdate)
	if err != nil {
		return errors.Wrapf(err, "while updating %q EventActivations", toUpdate.Name)
	}

	return nil
}

func (sm *migrationManager) backupOwnerRefFromEventActivation(ea *v1alpha1.EventActivation) error {
	instanceName, found := serviceInstanceKey(ea)
	if !found {
		return nil
	}

	cm, err := sm.k8sClient.CoreV1().ConfigMaps(sm.workingNS).Get(configMapNameForOwnerRef, metav1.GetOptions{})
	if err != nil {
		return err
	}

	cmCopy := cm.DeepCopy()
	if cmCopy.Data == nil {
		cmCopy.Data = map[string]string{}
	}

	cmCopy.Data[eventActivationKey(ea)] = instanceName

	_, err = sm.k8sClient.CoreV1().ConfigMaps(sm.workingNS).Update(cm)
	if err != nil {
		return err
	}

	return nil
}

func eventActivationKey(ea *v1alpha1.EventActivation) string {
	return fmt.Sprintf("%s/%s", ea.Namespace, ea.Name)
}

func serviceInstanceKey(ea *v1alpha1.EventActivation) (string, bool) {
	var instanceName = ""
	for _, oRef := range ea.OwnerReferences {
		if oRef.Kind == "ServiceInstance" {
			instanceName = oRef.Name
		}
	}
	if instanceName == "" {
		return "", false
	}

	return instanceName, true
}

func (sm *migrationManager) ensureBackupConfigMapExists() error {
	_, err := sm.k8sClient.CoreV1().ConfigMaps(sm.workingNS).Get(configMapNameForOwnerRef, metav1.GetOptions{})
	switch {
	case err == nil:
	case apiErr.IsNotFound(err):
		cm := v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      configMapNameForOwnerRef,
				Namespace: sm.workingNS,
			},
			Data: map[string]string{},
		}

		_, err := sm.k8sClient.CoreV1().ConfigMaps(sm.workingNS).Create(&cm)
		if err != nil {
			return fmt.Errorf("while creating ConfigMap %s/%s : %s", sm.workingNS, configMapNameForOwnerRef, err)
		}
	default:
		return fmt.Errorf("while getting ConfigMap %s/%s : %s", sm.workingNS, configMapNameForOwnerRef, err)
	}

	return nil
}
