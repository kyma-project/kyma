package kubernetes

import (
	"context"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var _ = ginkgo.Describe("Namespace", func() {
	var (
		reconciler         *NamespaceReconciler
		request            ctrl.Request
		baseSecret         *corev1.Secret
		baseConfigMap      *corev1.ConfigMap
		baseServiceAccount *corev1.ServiceAccount
		namespace          string
	)

	ginkgo.BeforeEach(func() {
		userNamespace := newFixNamespace("tam")
		gomega.Expect(resourceClient.Create(context.TODO(), userNamespace)).To(gomega.Succeed())

		baseConfigMap = newFixBaseConfigMap(config.BaseNamespace, "ah-tak-przeciez")
		gomega.Expect(resourceClient.Create(context.TODO(), baseConfigMap)).To(gomega.Succeed())

		baseSecret = newFixBaseSecret(config.BaseNamespace, "ah-tak-przeciez")
		gomega.Expect(resourceClient.Create(context.TODO(), baseSecret)).To(gomega.Succeed())

		baseServiceAccount = newFixBaseServiceAccount(config.BaseNamespace, "ah-tak-przeciez")
		gomega.Expect(resourceClient.Create(context.TODO(), baseServiceAccount)).To(gomega.Succeed())

		request = ctrl.Request{NamespacedName: types.NamespacedName{Name: userNamespace.GetName()}}
		reconciler = NewNamespace(k8sClient, log.Log, config, configMapSvc, secretSvc, serviceAccountSvc)
		namespace = userNamespace.GetName()
	})

	ginkgo.It("should successfully propagate base ServiceAccount to user namespace", func() {
		ginkgo.By("reconciling the Namespace")
		result, err := reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(0 * time.Second))

		configMap := &corev1.ConfigMap{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseConfigMap.GetName()}, configMap)).To(gomega.Succeed())
		compareConfigMaps(configMap, baseConfigMap)

		secret := &corev1.Secret{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseSecret.GetName()}, secret)).To(gomega.Succeed())
		compareSecrets(secret, baseSecret)

		serviceAccount := &corev1.ServiceAccount{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseServiceAccount.GetName()}, serviceAccount)).To(gomega.Succeed())
		compareServiceAccounts(serviceAccount, baseServiceAccount)

		ginkgo.By("one more time reconciling the Namespace")
		result, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(0 * time.Second))

		configMap = &corev1.ConfigMap{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseConfigMap.GetName()}, configMap)).To(gomega.Succeed())
		compareConfigMaps(configMap, baseConfigMap)

		secret = &corev1.Secret{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseSecret.GetName()}, secret)).To(gomega.Succeed())
		compareSecrets(secret, baseSecret)

		serviceAccount = &corev1.ServiceAccount{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseServiceAccount.GetName()}, serviceAccount)).To(gomega.Succeed())
		compareServiceAccounts(serviceAccount, baseServiceAccount)
	})
})
