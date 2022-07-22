package kubernetes

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"k8s.io/client-go/kubernetes/scheme"

	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
	"github.com/kyma-project/kyma/components/function-controller/internal/resource/automock"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestConfigMapReconciler_Reconcile(t *testing.T) {
	//GIVEN
	g := gomega.NewGomegaWithT(t)
	k8sClient, testEnv := setUpTestEnv(g)
	defer tearDownTestEnv(g, testEnv)
	resourceClient := resource.New(k8sClient, scheme.Scheme)
	testCfg := setUpControllerConfig(g)
	configMapSvc := NewConfigMapService(resourceClient, testCfg)

	baseNamespace := newFixNamespace(testCfg.BaseNamespace)
	g.Expect(k8sClient.Create(context.TODO(), baseNamespace)).To(gomega.Succeed())

	userNamespace := newFixNamespace("tam")
	g.Expect(k8sClient.Create(context.TODO(), userNamespace)).To(gomega.Succeed())

	baseConfigMap := newFixBaseConfigMap(testCfg.BaseNamespace, "ah-tak-przeciez")
	g.Expect(k8sClient.Create(context.TODO(), baseConfigMap)).To(gomega.Succeed())

	request := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: baseConfigMap.GetNamespace(), Name: baseConfigMap.GetName()}}
	reconciler := NewConfigMap(k8sClient, zap.NewNop().Sugar(), testCfg, configMapSvc)
	namespace := userNamespace.GetName()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//WHEN
	t.Log("reconciling ConfigMap that doesn't exist")
	_, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Namespace: baseConfigMap.GetNamespace(), Name: "not-existing-cm"}})
	g.Expect(err).To(gomega.BeNil(), "should not throw error on non existing configmap")

	t.Log("reconciling the ConfigMap")
	result, err := reconciler.Reconcile(ctx, request)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(result.Requeue).To(gomega.BeFalse())
	g.Expect(result.RequeueAfter).To(gomega.Equal(testCfg.ConfigMapRequeueDuration))

	configMap := &corev1.ConfigMap{}
	g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseConfigMap.GetName()}, configMap)).To(gomega.Succeed())
	compareConfigMaps(g, configMap, baseConfigMap)

	t.Log("updating the base ConfigMap")
	cmCopy := baseConfigMap.DeepCopy()
	cmCopy.Labels["test"] = "value"
	cmCopy.Data["test123"] = "321tset"
	g.Expect(k8sClient.Update(context.TODO(), cmCopy)).To(gomega.Succeed())

	result, err = reconciler.Reconcile(ctx, request)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(result.Requeue).To(gomega.BeFalse())
	g.Expect(result.RequeueAfter).To(gomega.Equal(testCfg.ConfigMapRequeueDuration))

	configMap = &corev1.ConfigMap{}
	g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseConfigMap.GetName()}, configMap)).To(gomega.Succeed())
	compareConfigMaps(g, configMap, cmCopy)

	t.Log("updating the modified ConfigMap in user namespace")
	userCopy := configMap.DeepCopy()
	userCopy.Data["4213"] = "142343"
	g.Expect(k8sClient.Update(context.TODO(), userCopy)).To(gomega.Succeed())

	result, err = reconciler.Reconcile(ctx, request)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(result.Requeue).To(gomega.BeFalse())
	g.Expect(result.RequeueAfter).To(gomega.Equal(testCfg.ConfigMapRequeueDuration))

	configMap = &corev1.ConfigMap{}
	g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseConfigMap.GetName()}, configMap)).To(gomega.Succeed())
	compareConfigMaps(g, configMap, cmCopy)
}

func TestConfigMapReconciler_predicate(t *testing.T) {
	baseNs := "base_ns"

	r := &ConfigMapReconciler{svc: NewConfigMapService(resource.New(&automock.K8sClient{}, runtime.NewScheme()), Config{BaseNamespace: baseNs})}
	preds := r.predicate()

	correctMeta := metav1.ObjectMeta{
		Namespace: baseNs,
		Labels:    map[string]string{ConfigLabel: RuntimeLabelValue},
	}

	pod := &corev1.Pod{ObjectMeta: correctMeta}
	labelledConfigmap := &corev1.ConfigMap{ObjectMeta: correctMeta}
	unlabelledConfigMap := &corev1.ConfigMap{}

	t.Run("deleteFunc", func(t *testing.T) {
		g := gomega.NewWithT(t)
		deleteEventPod := event.DeleteEvent{Object: pod}
		deleteEventPod.Object = pod
		deleteEventLabelledSrvAcc := event.DeleteEvent{Object: labelledConfigmap}
		deleteEventUnlabelledSrvAcc := event.DeleteEvent{Object: unlabelledConfigMap}

		g.Expect(preds.Delete(deleteEventPod)).To(gomega.BeFalse())
		g.Expect(preds.Delete(deleteEventLabelledSrvAcc)).To(gomega.BeFalse())
		g.Expect(preds.Delete(deleteEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		g.Expect(preds.Delete(event.DeleteEvent{})).To(gomega.BeFalse())
	})

	t.Run("createFunc", func(t *testing.T) {
		g := gomega.NewWithT(t)
		createEventPod := event.CreateEvent{Object: pod}
		createEventLabelledSrvAcc := event.CreateEvent{Object: labelledConfigmap}
		createEventUnlabelledSrvAcc := event.CreateEvent{Object: unlabelledConfigMap}

		g.Expect(preds.Create(createEventPod)).To(gomega.BeFalse())
		g.Expect(preds.Create(createEventLabelledSrvAcc)).To(gomega.BeTrue())
		g.Expect(preds.Create(createEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		g.Expect(preds.Create(event.CreateEvent{})).To(gomega.BeFalse())
	})

	t.Run("genericFunc", func(t *testing.T) {
		g := gomega.NewWithT(t)
		genericEventPod := event.GenericEvent{Object: pod}
		genericEventLabelledSrvAcc := event.GenericEvent{Object: labelledConfigmap}
		genericEventUnlabelledSrvAcc := event.GenericEvent{Object: unlabelledConfigMap}

		g.Expect(preds.Generic(genericEventPod)).To(gomega.BeFalse())
		g.Expect(preds.Generic(genericEventLabelledSrvAcc)).To(gomega.BeTrue())
		g.Expect(preds.Generic(genericEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		g.Expect(preds.Generic(event.GenericEvent{})).To(gomega.BeFalse())
	})

	t.Run("updateFunc", func(t *testing.T) {
		g := gomega.NewWithT(t)
		updateEventPod := event.UpdateEvent{ObjectNew: pod}
		updateEventLabelledSrvAcc := event.UpdateEvent{ObjectNew: labelledConfigmap}
		updateEventUnlabelledSrvAcc := event.UpdateEvent{ObjectNew: unlabelledConfigMap}

		g.Expect(preds.Update(updateEventPod)).To(gomega.BeFalse())
		g.Expect(preds.Update(updateEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		g.Expect(preds.Update(updateEventLabelledSrvAcc)).To(gomega.BeTrue())
		g.Expect(preds.Update(event.UpdateEvent{})).To(gomega.BeFalse())
	})
}
