package kubernetes

import (
	"context"
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
	"github.com/kyma-project/kyma/components/function-controller/internal/resource/automock"
)

var _ = ginkgo.Describe("ConfigMap", func() {

	var (
		reconciler    *ConfigMapReconciler
		request       ctrl.Request
		baseConfigMap *corev1.ConfigMap
		namespace     string
	)

	ginkgo.BeforeEach(func() {
		userNamespace := newFixNamespace("tam")
		gomega.Expect(resourceClient.Create(context.TODO(), userNamespace)).To(gomega.Succeed())

		baseConfigMap = newFixBaseConfigMap(config.BaseNamespace, "ah-tak-przeciez")
		gomega.Expect(resourceClient.Create(context.TODO(), baseConfigMap)).To(gomega.Succeed())

		request = ctrl.Request{NamespacedName: types.NamespacedName{Namespace: baseConfigMap.GetNamespace(), Name: baseConfigMap.GetName()}}
		reconciler = NewConfigMap(k8sClient, log.Log, config, configMapSvc)
		namespace = userNamespace.GetName()
	})

	ginkgo.It("should successfully propagate base ConfigMap to user namespaceeeee", func() {
		ginkgo.By("reconciling ConfigMap that doesn't exist")
		_, err := reconciler.Reconcile(ctrl.Request{NamespacedName: types.NamespacedName{Namespace: baseConfigMap.GetNamespace(), Name: "not-existing-cm"}})
		gomega.Expect(err).To(gomega.BeNil(), "should not throw error on non existing configmap")

		ginkgo.By("reconciling the ConfigMap")
		result, err := reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(config.ConfigMapRequeueDuration))

		configMap := &corev1.ConfigMap{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseConfigMap.GetName()}, configMap)).To(gomega.Succeed())
		compareConfigMaps(configMap, baseConfigMap)

		ginkgo.By("updating the base ConfigMap")
		copy := baseConfigMap.DeepCopy()
		copy.Labels["test"] = "value"
		copy.Data["test123"] = "321tset"
		gomega.Expect(k8sClient.Update(context.TODO(), copy)).To(gomega.Succeed())

		result, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(config.ConfigMapRequeueDuration))

		configMap = &corev1.ConfigMap{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseConfigMap.GetName()}, configMap)).To(gomega.Succeed())
		compareConfigMaps(configMap, copy)

		ginkgo.By("updating the modified ConfigMap in user namespace")
		userCopy := configMap.DeepCopy()
		userCopy.Data["4213"] = "142343"
		gomega.Expect(k8sClient.Update(context.TODO(), userCopy)).To(gomega.Succeed())

		result, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(config.ConfigMapRequeueDuration))

		configMap = &corev1.ConfigMap{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseConfigMap.GetName()}, configMap)).To(gomega.Succeed())
		compareConfigMaps(configMap, copy)
	})
})

func TestConfigMapReconciler_predicate(t *testing.T) {
	gm := gomega.NewGomegaWithT(t)
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
		deleteEventPod := event.DeleteEvent{Meta: pod.GetObjectMeta(), Object: pod}
		deleteEventLabelledSrvAcc := event.DeleteEvent{Meta: labelledConfigmap.GetObjectMeta(), Object: labelledConfigmap}
		deleteEventUnlabelledSrvAcc := event.DeleteEvent{Meta: unlabelledConfigMap.GetObjectMeta(), Object: unlabelledConfigMap}

		gm.Expect(preds.Delete(deleteEventPod)).To(gomega.BeFalse())
		gm.Expect(preds.Delete(deleteEventLabelledSrvAcc)).To(gomega.BeFalse())
		gm.Expect(preds.Delete(deleteEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		gm.Expect(preds.Delete(event.DeleteEvent{})).To(gomega.BeFalse())
	})

	t.Run("createFunc", func(t *testing.T) {
		createEventPod := event.CreateEvent{Meta: pod.GetObjectMeta(), Object: pod}
		createEventLabelledSrvAcc := event.CreateEvent{Meta: labelledConfigmap.GetObjectMeta(), Object: labelledConfigmap}
		createEventUnlabelledSrvAcc := event.CreateEvent{Meta: unlabelledConfigMap.GetObjectMeta(), Object: unlabelledConfigMap}

		gm.Expect(preds.Create(createEventPod)).To(gomega.BeFalse())
		gm.Expect(preds.Create(createEventLabelledSrvAcc)).To(gomega.BeTrue())
		gm.Expect(preds.Create(createEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		gm.Expect(preds.Create(event.CreateEvent{})).To(gomega.BeFalse())
	})

	t.Run("genericFunc", func(t *testing.T) {
		genericEventPod := event.GenericEvent{Meta: pod.GetObjectMeta(), Object: pod}
		genericEventLabelledSrvAcc := event.GenericEvent{Meta: labelledConfigmap.GetObjectMeta(), Object: labelledConfigmap}
		genericEventUnlabelledSrvAcc := event.GenericEvent{Meta: unlabelledConfigMap.GetObjectMeta(), Object: unlabelledConfigMap}

		gm.Expect(preds.Generic(genericEventPod)).To(gomega.BeFalse())
		gm.Expect(preds.Generic(genericEventLabelledSrvAcc)).To(gomega.BeTrue())
		gm.Expect(preds.Generic(genericEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		gm.Expect(preds.Generic(event.GenericEvent{})).To(gomega.BeFalse())
	})

	t.Run("updateFunc", func(t *testing.T) {
		updateEventPod := event.UpdateEvent{MetaNew: pod.GetObjectMeta(), ObjectNew: pod}
		updateEventLabelledSrvAcc := event.UpdateEvent{MetaNew: labelledConfigmap.GetObjectMeta(), ObjectNew: labelledConfigmap}
		updateEventUnlabelledSrvAcc := event.UpdateEvent{MetaNew: unlabelledConfigMap.GetObjectMeta(), ObjectNew: unlabelledConfigMap}

		gm.Expect(preds.Update(updateEventPod)).To(gomega.BeFalse())
		gm.Expect(preds.Update(updateEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		gm.Expect(preds.Update(updateEventLabelledSrvAcc)).To(gomega.BeTrue())
		gm.Expect(preds.Update(event.UpdateEvent{})).To(gomega.BeFalse())
	})
}
