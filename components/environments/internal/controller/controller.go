package controller

import (
	"fmt"
	"log"

	"github.com/kyma-project/kyma/components/environments/internal"
	"github.com/kyma-project/kyma/components/environments/internal/limit_range"
	"github.com/kyma-project/kyma/components/environments/internal/namespaces"
	rq "github.com/kyma-project/kyma/components/environments/internal/resource-quota"
	"github.com/kyma-project/kyma/components/environments/internal/roles"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	rolesAnnotName           = "kyma-roles"
	labelSelector            = "env=true"
	istioInjectionLabel      = "istio-injection"
	istioInjectionLabelValue = "enabled"
	resourceQuotaName        = "kyma-default"
	limitRangeName           = "kyma-default"
)

var listOptions = metav1.ListOptions{LabelSelector: labelSelector}

type environments struct {
	Clientset           *kubernetes.Clientset
	NamespacesClient    namespaces.NamespacesClientInterface
	RolesClient         roles.RolesClientInterface
	LimitRangeClient    limit_range.LimitRangesClientInterface
	ResourceQuotaClient rq.ResourceQuotaClientInterface
	Config              *EnvironmentsConfig
	ErrorHandlers       internal.ErrorHandlersInterface
}

func NewController(clientset *kubernetes.Clientset, config *EnvironmentsConfig) (cache.Controller, error) {
	envs := environments{
		Clientset:           clientset,
		Config:              config,
		NamespacesClient:    &namespaces.NamespacesClient{Clientset: clientset},
		RolesClient:         &roles.RolesClient{Clientset: clientset},
		LimitRangeClient:    &limit_range.LimitRangeClient{Clientset: clientset},
		ResourceQuotaClient: &rq.ResourceQuotaClient{Clientset: clientset},
		ErrorHandlers:       &internal.ErrorHandlers{},
	}

	_, controller := cache.NewInformer(
		enviromentsWatcher{clientset.CoreV1().Namespaces()},
		&v1.Namespace{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    envs.onAdd,
			UpdateFunc: func(oldObj, newObj interface{}) {},
			DeleteFunc: envs.onDelete,
		})

	return controller, nil
}

func (envs *environments) onAdd(obj interface{}) {
	namespace := obj.(*v1.Namespace)

	log.Printf("[CONTROLLER] onAdd triggered for %s\n", namespace.Name)

	envs.AddRolesForEnvironment(namespace)
	envs.LabelWithIstioInjection(namespace)
	envs.CreateLimitRangeForEnv(namespace)
	envs.CreateResourceQuota(namespace)
}

func (envs *environments) onDelete(obj interface{}) {
	namespace := obj.(*v1.Namespace)

	log.Printf("[CONTROLLER] onDelete triggered for %s\n", namespace.Name)

	envs.RemoveRolesFromEnvironment(namespace)
	envs.RemoveIstioInjectionLabel(namespace)
	envs.DeleteLimitRange(namespace)
	envs.DeleteResourceQuota(namespace)
}

func hasRoles(namespace *v1.Namespace) bool {
	return contains(namespace.ObjectMeta.Annotations, rolesAnnotName)
}

func (envs *environments) AddRolesForEnvironment(environment *v1.Namespace) error {
	namespace, err := envs.NamespacesClient.GetNamespace(environment.Name)
	if envs.ErrorHandlers.CheckError("Error while getting namespace.", err) {
		return err
	}

	if hasRoles(namespace) {
		log.Println("Namespace already have default roles.")
		return nil
	}

	rolesToBootstrap, err := envs.RolesClient.GetList(envs.Config.Namespace, listOptions)
	if envs.ErrorHandlers.CheckError("Error on fetching roles to bootstrap.", err) {
		return err
	}

	for _, role := range rolesToBootstrap.Items {
		roleCopy := role.DeepCopy()
		roleCopy.ObjectMeta = metav1.ObjectMeta{
			Name:      role.ObjectMeta.Name,
			Namespace: namespace.Name,
		}

		_, err = envs.RolesClient.CreateRole(roleCopy, namespace.Name)
		envs.ErrorHandlers.LogError(fmt.Sprintf("Error on creating %s role", roleCopy.ObjectMeta.Name), err)
	}

	err = envs.annotateObject(namespace, rolesAnnotName)
	if envs.ErrorHandlers.CheckError("Error on updating namespace.", err) {
		return err
	}

	return nil
}

func (envs *environments) LabelWithIstioInjection(environment *v1.Namespace) error {
	namespace, err := envs.NamespacesClient.GetNamespace(environment.Name)
	if envs.ErrorHandlers.CheckError("Error while getting namespace.", err) {
		return err
	}

	err = envs.labelNamespace(namespace, istioInjectionLabel, istioInjectionLabelValue)
	if envs.ErrorHandlers.CheckError("Error on updating namespace.", err) {
		return err
	}

	return nil
}

func (envs *environments) CreateLimitRangeForEnv(environment *v1.Namespace) error {
	err := envs.LimitRangeClient.CreateLimitRange(environment.Name, &v1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name: limitRangeName,
		},
		Spec: v1.LimitRangeSpec{
			Limits: []v1.LimitRangeItem{
				{
					Type: v1.LimitTypeContainer,
					Default: v1.ResourceList{
						v1.ResourceMemory: *envs.Config.LimitRangeMemory.Default.AsQuantity(),
					},
					DefaultRequest: v1.ResourceList{
						v1.ResourceMemory: *envs.Config.LimitRangeMemory.DefaultRequest.AsQuantity(),
					},
					Max: v1.ResourceList{
						v1.ResourceMemory: *envs.Config.LimitRangeMemory.Max.AsQuantity(),
					},
				},
			},
		},
	})
	if envs.ErrorHandlers.CheckError("Error on creating limit range", err) {
		return err
	}

	return nil
}

func (envs *environments) CreateResourceQuota(namespace *v1.Namespace) error {
	err := envs.ResourceQuotaClient.CreateResourceQuota(namespace.Name, &v1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name: resourceQuotaName,
		},
		Spec: v1.ResourceQuotaSpec{
			Hard: v1.ResourceList{
				v1.ResourceRequestsMemory: *envs.Config.ResourceQuota.RequestsMemory.AsQuantity(),
				v1.ResourceLimitsMemory:   *envs.Config.ResourceQuota.LimitsMemory.AsQuantity(),
			},
		},
	})

	if envs.ErrorHandlers.CheckError("Error on creating resource quota", err) {
		return err
	}

	return nil
}

func (envs *environments) annotateObject(namespace *v1.Namespace, annotName string) error {
	origAnnotations := namespace.GetAnnotations()

	if origAnnotations == nil {
		origAnnotations = make(map[string]string)
	}

	origAnnotations[annotName] = "true"

	namespaceCopy := namespace.DeepCopy()
	namespaceCopy.SetAnnotations(origAnnotations)

	_, err := envs.NamespacesClient.UpdateNamespace(namespaceCopy)

	if envs.ErrorHandlers.CheckError("Error annotating object", err) {
		return err
	}

	return nil
}

func (envs *environments) labelNamespace(namespace *v1.Namespace, labelName string, labelValue string) error {
	origLabels := namespace.GetLabels()

	origLabels[labelName] = labelValue

	namespaceCopy := namespace.DeepCopy()
	namespaceCopy.SetLabels(origLabels)

	_, err := envs.NamespacesClient.UpdateNamespace(namespaceCopy)
	if envs.ErrorHandlers.CheckError("Error labelling object", err) {
		return err
	}

	return nil
}

func (envs *environments) RemoveRolesFromEnvironment(environment *v1.Namespace) error {
	namespace, err := envs.NamespacesClient.GetNamespace(environment.Name)

	if envs.ErrorHandlers.CheckError("Error while getting namespace.", err) {
		return err
	}

	if !hasRoles(namespace) {
		return nil
	}

	err = envs.unannotateObject(namespace, rolesAnnotName)
	if envs.ErrorHandlers.CheckError("Error on updating namespace.", err) {
		return err
	}

	rolesToDelete, err := envs.RolesClient.GetList(envs.Config.Namespace, listOptions)
	if envs.ErrorHandlers.CheckError("Error on fetching roles to delete.", err) {
		return err
	}

	for _, role := range rolesToDelete.Items {
		err = envs.RolesClient.DeleteRole(role.ObjectMeta.Name, namespace.Name)
		envs.ErrorHandlers.LogError(fmt.Sprintf("Error on deleting %s role", role.ObjectMeta.Name), err)
	}

	return nil
}

func (envs *environments) RemoveIstioInjectionLabel(environment *v1.Namespace) error {

	namespace, err := envs.NamespacesClient.GetNamespace(environment.Name)
	if envs.ErrorHandlers.CheckError("Error while getting namespace.", err) {
		return err
	}

	err = envs.removeLabelFromNamespace(namespace, istioInjectionLabel)
	if envs.ErrorHandlers.CheckError("Error on updating namespace.", err) {
		return err
	}

	return nil
}

func (envs *environments) DeleteLimitRange(environment *v1.Namespace) error {
	err := envs.LimitRangeClient.DeleteLimitRange(environment.Name)
	if envs.ErrorHandlers.CheckError("Error on deleting limit range.", err) {
		return err
	}

	return nil
}

func (envs *environments) DeleteResourceQuota(namespace *v1.Namespace) error {
	err := envs.ResourceQuotaClient.DeleteResourceQuota(namespace.Name)
	if envs.ErrorHandlers.CheckError("Error on deleting resource quota.", err) {
		return err
	}

	return nil
}

func (envs *environments) unannotateObject(namespace *v1.Namespace, annotName string) error {
	origAnnotations := namespace.GetAnnotations()

	if origAnnotations == nil {
		log.Println("Unable to unannotate. Provided object is not annotated!")
		return nil
	}

	delete(origAnnotations, annotName)

	namespaceCopy := namespace.DeepCopy()
	namespaceCopy.SetAnnotations(origAnnotations)

	_, err := envs.NamespacesClient.UpdateNamespace(namespaceCopy)
	if envs.ErrorHandlers.CheckError("Error unannotating object", err) {
		return err
	}

	return nil
}

func (envs *environments) removeLabelFromNamespace(namespace *v1.Namespace, labelName string) error {
	origLabels := namespace.GetLabels()

	delete(origLabels, labelName)

	namespaceCopy := namespace.DeepCopy()
	namespaceCopy.SetLabels(origLabels)

	_, err := envs.NamespacesClient.UpdateNamespace(namespaceCopy)
	if envs.ErrorHandlers.CheckError("Error removing label", err) {
		return err
	}

	return nil
}

func contains(slice map[string]string, item string) bool {
	for key := range slice {
		if key == item {
			return true
		}
	}

	return false
}
