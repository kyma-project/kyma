package controller

import (
	"log"

	"github.com/kyma-project/kyma/components/namespace-controller/internal"
	"github.com/kyma-project/kyma/components/namespace-controller/internal/limit_range"
	"github.com/kyma-project/kyma/components/namespace-controller/internal/namespaces"
	rq "github.com/kyma-project/kyma/components/namespace-controller/internal/resource-quota"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	labelSelector            = "env=true"
	istioInjectionLabel      = "istio-injection"
	istioInjectionLabelValue = "enabled"
	resourceQuotaName        = "kyma-default"
	limitRangeName           = "kyma-default"
)

var listOptions = metav1.ListOptions{LabelSelector: labelSelector}

type controller struct {
	Clientset           *kubernetes.Clientset
	NamespacesClient    namespaces.NamespacesClientInterface
	LimitRangeClient    limit_range.LimitRangesClientInterface
	ResourceQuotaClient rq.ResourceQuotaClientInterface
	Config              *NamespacesConfig
	ErrorHandlers       internal.ErrorHandlersInterface
}

func NewController(clientset *kubernetes.Clientset, config *NamespacesConfig) (cache.Controller, error) {
	envs := controller{
		Clientset:           clientset,
		Config:              config,
		NamespacesClient:    &namespaces.NamespacesClient{Clientset: clientset},
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

func (c *controller) onAdd(obj interface{}) {
	namespace := obj.(*v1.Namespace)

	log.Printf("[CONTROLLER] onAdd triggered for %s\n", namespace.Name)

	err := c.LabelWithIstioInjection(namespace)
	if err != nil {
		log.Printf("Cannot label namespace '%s' with istio injection: %v", namespace, err)
	}
	err = c.CreateLimitRangeForEnv(namespace)
	if err != nil {
		log.Printf("Cannot create limit range for namespace '%s': %v", namespace, err)
	}
	err = c.CreateResourceQuota(namespace)
	if err != nil {
		log.Printf("Cannot create resource quota for namespace '%s': %v", namespace, err)
	}
}

func (c *controller) onDelete(obj interface{}) {
	namespace := obj.(*v1.Namespace)

	log.Printf("[CONTROLLER] onDelete triggered for %s\n", namespace.Name)

	err := c.RemoveIstioInjectionLabel(namespace)
	if err != nil {
		log.Printf("Cannot remove istio injection label from namespace '%s': %v", namespace, err)
	}
	err = c.DeleteLimitRange(namespace)
	if err != nil {
		log.Printf("Cannot delete limit range for namespace '%s': %v", namespace, err)
	}
	err = c.DeleteResourceQuota(namespace)
	if err != nil {
		log.Printf("Cannot delete resource quota for namespace '%s': %v", namespace, err)
	}
}

func (c *controller) LabelWithIstioInjection(namespace *v1.Namespace) error {
	namespace, err := c.NamespacesClient.GetNamespace(namespace.Name)
	if c.ErrorHandlers.CheckError("Error while getting namespace.", err) {
		return err
	}

	err = c.labelNamespace(namespace, istioInjectionLabel, istioInjectionLabelValue)
	if c.ErrorHandlers.CheckError("Error on updating namespace.", err) {
		return err
	}

	return nil
}

func (c *controller) CreateLimitRangeForEnv(namespace *v1.Namespace) error {
	err := c.LimitRangeClient.CreateLimitRange(namespace.Name, &v1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{
			Name: limitRangeName,
		},
		Spec: v1.LimitRangeSpec{
			Limits: []v1.LimitRangeItem{
				{
					Type: v1.LimitTypeContainer,
					Default: v1.ResourceList{
						v1.ResourceMemory: *c.Config.LimitRangeMemory.Default.AsQuantity(),
					},
					DefaultRequest: v1.ResourceList{
						v1.ResourceMemory: *c.Config.LimitRangeMemory.DefaultRequest.AsQuantity(),
					},
					Max: v1.ResourceList{
						v1.ResourceMemory: *c.Config.LimitRangeMemory.Max.AsQuantity(),
					},
				},
			},
		},
	})
	if c.ErrorHandlers.CheckError("Error on creating limit range", err) {
		return err
	}

	return nil
}

func (c *controller) CreateResourceQuota(namespace *v1.Namespace) error {
	err := c.ResourceQuotaClient.CreateResourceQuota(namespace.Name, &v1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name: resourceQuotaName,
		},
		Spec: v1.ResourceQuotaSpec{
			Hard: v1.ResourceList{
				v1.ResourceRequestsMemory: *c.Config.ResourceQuota.RequestsMemory.AsQuantity(),
				v1.ResourceLimitsMemory:   *c.Config.ResourceQuota.LimitsMemory.AsQuantity(),
			},
		},
	})

	if c.ErrorHandlers.CheckError("Error on creating resource quota", err) {
		return err
	}

	return nil
}

func (c *controller) annotateObject(namespace *v1.Namespace, annotName string) error {
	origAnnotations := namespace.GetAnnotations()

	if origAnnotations == nil {
		origAnnotations = make(map[string]string)
	}

	origAnnotations[annotName] = "true"

	namespaceCopy := namespace.DeepCopy()
	namespaceCopy.SetAnnotations(origAnnotations)

	_, err := c.NamespacesClient.UpdateNamespace(namespaceCopy)

	if c.ErrorHandlers.CheckError("Error annotating object", err) {
		return err
	}

	return nil
}

func (c *controller) labelNamespace(namespace *v1.Namespace, labelName string, labelValue string) error {
	origLabels := namespace.GetLabels()

	origLabels[labelName] = labelValue

	namespaceCopy := namespace.DeepCopy()
	namespaceCopy.SetLabels(origLabels)

	_, err := c.NamespacesClient.UpdateNamespace(namespaceCopy)
	if c.ErrorHandlers.CheckError("Error labelling object", err) {
		return err
	}

	return nil
}

func (c *controller) RemoveIstioInjectionLabel(namespace *v1.Namespace) error {

	namespace, err := c.NamespacesClient.GetNamespace(namespace.Name)
	if c.ErrorHandlers.CheckError("Error while getting namespace.", err) {
		return err
	}

	err = c.removeLabelFromNamespace(namespace, istioInjectionLabel)
	if c.ErrorHandlers.CheckError("Error on updating namespace.", err) {
		return err
	}

	return nil
}

func (c *controller) DeleteLimitRange(namespace *v1.Namespace) error {
	err := c.LimitRangeClient.DeleteLimitRange(namespace.Name)
	if c.ErrorHandlers.CheckError("Error on deleting limit range.", err) {
		return err
	}

	return nil
}

func (c *controller) DeleteResourceQuota(namespace *v1.Namespace) error {
	err := c.ResourceQuotaClient.DeleteResourceQuota(namespace.Name)
	if c.ErrorHandlers.CheckError("Error on deleting resource quota.", err) {
		return err
	}

	return nil
}

func (c *controller) unannotateObject(namespace *v1.Namespace, annotName string) error {
	origAnnotations := namespace.GetAnnotations()

	if origAnnotations == nil {
		log.Println("Unable to unannotate. Provided object is not annotated!")
		return nil
	}

	delete(origAnnotations, annotName)

	namespaceCopy := namespace.DeepCopy()
	namespaceCopy.SetAnnotations(origAnnotations)

	_, err := c.NamespacesClient.UpdateNamespace(namespaceCopy)
	if c.ErrorHandlers.CheckError("Error unannotating object", err) {
		return err
	}

	return nil
}

func (c *controller) removeLabelFromNamespace(namespace *v1.Namespace, labelName string) error {
	origLabels := namespace.GetLabels()

	delete(origLabels, labelName)

	namespaceCopy := namespace.DeepCopy()
	namespaceCopy.SetLabels(origLabels)

	_, err := c.NamespacesClient.UpdateNamespace(namespaceCopy)
	if c.ErrorHandlers.CheckError("Error removing label", err) {
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
