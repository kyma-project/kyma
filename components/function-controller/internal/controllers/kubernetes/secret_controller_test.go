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

var _ = ginkgo.Describe("Secret", func() {
	var (
		reconciler *SecretReconciler
		request    ctrl.Request
		baseSecret *corev1.Secret
		namespace  string
	)

	ginkgo.BeforeEach(func() {
		userNamespace := newFixNamespace("tam")
		gomega.Expect(resourceClient.Create(context.TODO(), userNamespace)).To(gomega.Succeed())

		baseSecret = newFixBaseSecret(config.BaseNamespace, "ah-tak-przeciez")
		gomega.Expect(resourceClient.Create(context.TODO(), baseSecret)).To(gomega.Succeed())

		request = ctrl.Request{NamespacedName: types.NamespacedName{Namespace: baseSecret.GetNamespace(), Name: baseSecret.GetName()}}
		reconciler = NewSecret(k8sClient, log.Log, config, secretSvc)
		namespace = userNamespace.GetName()
	})

	ginkgo.It("should successfully propagate base Secret to user namespace", func() {
		ginkgo.By("reconciling non-existing secret")
		_, err := reconciler.Reconcile(ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: baseSecret.GetNamespace(),
				Name:      "not-existing-secret",
			},
		})
		gomega.Expect(err).To(gomega.BeNil())

		ginkgo.By("reconciling the Secret")
		result, err := reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(config.SecretRequeueDuration))

		secret := &corev1.Secret{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseSecret.GetName()}, secret)).To(gomega.Succeed())
		compareSecrets(secret, baseSecret)

		ginkgo.By("updating the base Secret")
		copy := baseSecret.DeepCopy()
		copy.Labels["test"] = "value"
		copy.Data["test123"] = []byte("321tset")
		gomega.Expect(k8sClient.Update(context.TODO(), copy)).To(gomega.Succeed())

		result, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(config.SecretRequeueDuration))

		secret = &corev1.Secret{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseSecret.GetName()}, secret)).To(gomega.Succeed())
		compareSecrets(secret, copy)

		ginkgo.By("updating the modified Secret in user namespace")
		userCopy := secret.DeepCopy()
		userCopy.Data["test123"] = []byte("321tset")
		gomega.Expect(k8sClient.Update(context.TODO(), userCopy)).To(gomega.Succeed())

		result, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(config.SecretRequeueDuration))

		secret = &corev1.Secret{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseSecret.GetName()}, secret)).To(gomega.Succeed())
		compareSecrets(secret, copy)
	})
})
