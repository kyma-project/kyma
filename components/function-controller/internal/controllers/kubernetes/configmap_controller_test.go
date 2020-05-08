package kubernetes

import (
	"context"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
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

	ginkgo.It("should successfully propagate base ConfigMap to user namespace", func() {
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
