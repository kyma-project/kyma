package kubernetes

import (
	"context"
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
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

		updatedBase := &corev1.Secret{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: baseSecret.GetNamespace(), Name: baseSecret.GetName()}, updatedBase)).To(gomega.Succeed())
		gomega.Expect(updatedBase.Finalizers).To(gomega.ContainElement(cfgSecretFinalizerName), "created base secret should have finalizer applied")
		secret := &corev1.Secret{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseSecret.GetName()}, secret)).To(gomega.Succeed())
		compareSecrets(secret, baseSecret)

		ginkgo.By("updating the base Secret")
		copy := updatedBase.DeepCopy()
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

	ginkgo.It("should successfully delete propagated Secrets from user namespace when base Secret is deleted", func() {
		ginkgo.By("reconciling the Secret")
		result, err := reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(config.SecretRequeueDuration))

		updatedBase := &corev1.Secret{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: baseSecret.GetNamespace(), Name: baseSecret.GetName()}, updatedBase)).To(gomega.Succeed())
		gomega.Expect(updatedBase.Finalizers).To(gomega.ContainElement(cfgSecretFinalizerName), "created base secret should have finalizer applied")
		secret := &corev1.Secret{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseSecret.GetName()}, secret)).To(gomega.Succeed())
		compareSecrets(secret, baseSecret)

		ginkgo.By("deleting base Secret")
		gomega.Expect(k8sClient.Delete(context.TODO(), updatedBase)).To(gomega.Succeed())
		result, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.BeZero())
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: updatedBase.GetNamespace(), Name: updatedBase.GetName()}, updatedBase)).To(gomega.HaveOccurred())
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: secret.GetNamespace(), Name: secret.GetName()}, secret)).To(gomega.HaveOccurred())
	})
})

func TestSecretReconciler_predicate(t *testing.T) {
	gm := gomega.NewGomegaWithT(t)

	baseSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: "base-ns", Labels: map[string]string{ConfigLabel: CredentialsLabelValue}}}
	nonBaseSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: "some-other-ns"}}
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod-name"}}

	r := &SecretReconciler{svc: &secretService{
		config: Config{
			BaseNamespace: "base-ns",
		},
	}}
	preds := r.predicate()

	t.Run("deleteFunc", func(t *testing.T) {
		podEvent := event.DeleteEvent{Meta: pod.GetObjectMeta(), Object: pod}
		eventBaseSecret := event.DeleteEvent{Meta: baseSecret.GetObjectMeta(), Object: baseSecret}
		eventNonBaseSecret := event.DeleteEvent{Meta: nonBaseSecret.GetObjectMeta(), Object: nonBaseSecret}

		gm.Expect(preds.Delete(podEvent)).To(gomega.BeFalse())
		gm.Expect(preds.Delete(eventBaseSecret)).To(gomega.BeTrue(), "should be true for base secret")
		gm.Expect(preds.Delete(eventNonBaseSecret)).To(gomega.BeFalse())
		gm.Expect(preds.Delete(event.DeleteEvent{})).To(gomega.BeFalse())
	})

	t.Run("createFunc", func(t *testing.T) {
		podEvent := event.CreateEvent{Meta: pod.GetObjectMeta(), Object: pod}
		eventBaseSecret := event.CreateEvent{Meta: baseSecret.GetObjectMeta(), Object: baseSecret}
		eventNonBaseSecret := event.CreateEvent{Meta: nonBaseSecret.GetObjectMeta(), Object: nonBaseSecret}

		gm.Expect(preds.Create(podEvent)).To(gomega.BeFalse())
		gm.Expect(preds.Create(eventNonBaseSecret)).To(gomega.BeFalse())
		gm.Expect(preds.Create(event.CreateEvent{})).To(gomega.BeFalse())

		gm.Expect(preds.Create(eventBaseSecret)).To(gomega.BeTrue(), "should be true for base secret")
	})

	t.Run("genericFunc", func(t *testing.T) {
		podEvent := event.GenericEvent{Meta: pod.GetObjectMeta(), Object: pod}
		eventBaseSecret := event.GenericEvent{Meta: baseSecret.GetObjectMeta(), Object: baseSecret}
		eventNonBaseSecret := event.GenericEvent{Meta: nonBaseSecret.GetObjectMeta(), Object: nonBaseSecret}

		gm.Expect(preds.Generic(podEvent)).To(gomega.BeFalse())
		gm.Expect(preds.Generic(eventNonBaseSecret)).To(gomega.BeFalse())
		gm.Expect(preds.Generic(event.GenericEvent{})).To(gomega.BeFalse())

		gm.Expect(preds.Generic(eventBaseSecret)).To(gomega.BeTrue(), "should be true for base secret")
	})

	t.Run("updateFunc", func(t *testing.T) {
		podEvent := event.UpdateEvent{MetaNew: pod.GetObjectMeta(), ObjectNew: pod}
		eventBaseSecret := event.UpdateEvent{MetaNew: baseSecret.GetObjectMeta(), ObjectNew: baseSecret}
		eventNonBaseSecret := event.UpdateEvent{MetaNew: nonBaseSecret.GetObjectMeta(), ObjectNew: nonBaseSecret}

		gm.Expect(preds.Update(podEvent)).To(gomega.BeFalse())
		gm.Expect(preds.Update(eventNonBaseSecret)).To(gomega.BeFalse())
		gm.Expect(preds.Update(event.UpdateEvent{})).To(gomega.BeFalse())

		gm.Expect(preds.Update(eventBaseSecret)).To(gomega.BeTrue(), "should be true for base secret")
	})
}
