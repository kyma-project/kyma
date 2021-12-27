package kubernetes

import (
	"context"
	"testing"

	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func TestSecretReconciler_Reconcile(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)
	k8sClient, testEnv := setUpTestEnv(g)
	defer tearDownTestEnv(g, testEnv)
	resourceClient := resource.New(k8sClient, scheme.Scheme)
	testCfg := setUpControllerConfig(g)

	baseNamespace := newFixNamespace(testCfg.BaseNamespace)
	g.Expect(k8sClient.Create(context.TODO(), baseNamespace)).To(gomega.Succeed())

	userNamespace := newFixNamespace("tam")
	g.Expect(resourceClient.Create(context.TODO(), userNamespace)).To(gomega.Succeed())

	secretSvc := NewSecretService(resourceClient, testCfg)
	reconciler := NewSecret(k8sClient, log.Log, testCfg, secretSvc)

	namespace := userNamespace.GetName()

	t.Run("should successfully propagate base Secret to user namespace", func(t *testing.T) {
		//GIVEN
		g := gomega.NewGomegaWithT(t)
		t.Log("reconciling non-existing secret")
		baseSecret := newFixBaseSecret(testCfg.BaseNamespace, "successful-propagation")
		createSecret(g, resourceClient, baseSecret)
		defer deleteSecret(g, k8sClient, baseSecret)
		_, err := reconciler.Reconcile(ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: baseSecret.GetNamespace(),
				Name:      "not-existing-secret",
			},
		})
		g.Expect(err).To(gomega.BeNil())
		request := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: testCfg.BaseNamespace, Name: baseSecret.GetName()}}

		//WHEN
		t.Log("reconciling the Secret")
		result, err := reconciler.Reconcile(request)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.Requeue).To(gomega.BeFalse())
		g.Expect(result.RequeueAfter).To(gomega.Equal(testCfg.SecretRequeueDuration))

		updatedBase := &corev1.Secret{}
		g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: baseSecret.GetNamespace(), Name: baseSecret.GetName()}, updatedBase)).To(gomega.Succeed())
		g.Expect(updatedBase.Finalizers).To(gomega.ContainElement(cfgSecretFinalizerName), "created base secret should have finalizer applied")
		secret := &corev1.Secret{}
		g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseSecret.GetName()}, secret)).To(gomega.Succeed())
		compareSecrets(g, secret, baseSecret)

		t.Log("updating the base Secret")
		updateBaseSecretCopy := updatedBase.DeepCopy()
		updateBaseSecretCopy.Labels["test"] = "value"
		updateBaseSecretCopy.Data["test123"] = []byte("321tset")
		g.Expect(k8sClient.Update(context.TODO(), updateBaseSecretCopy)).To(gomega.Succeed())

		result, err = reconciler.Reconcile(request)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.Requeue).To(gomega.BeFalse())
		g.Expect(result.RequeueAfter).To(gomega.Equal(testCfg.SecretRequeueDuration))

		secret = &corev1.Secret{}
		g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseSecret.GetName()}, secret)).To(gomega.Succeed())
		compareSecrets(g, secret, updateBaseSecretCopy)

		t.Log("updating the modified Secret in user namespace")
		userCopy := secret.DeepCopy()
		userCopy.Data["test123"] = []byte("321tset")
		g.Expect(k8sClient.Update(context.TODO(), userCopy)).To(gomega.Succeed())

		result, err = reconciler.Reconcile(request)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.Requeue).To(gomega.BeFalse())
		g.Expect(result.RequeueAfter).To(gomega.Equal(testCfg.SecretRequeueDuration))

		secret = &corev1.Secret{}
		g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseSecret.GetName()}, secret)).To(gomega.Succeed())
		compareSecrets(g, secret, updateBaseSecretCopy)
	})

	t.Run("should not successfully propagate Secret managed by user to user namespace", func(t *testing.T) {
		//GIVEN
		g := gomega.NewGomegaWithT(t)
		t.Log("reconciling non-existing secret")

		baseSecretWithManagedLabel := newFixBaseSecretWithManagedLabel(testCfg.BaseNamespace, "secret-with-managed-label")
		g.Expect(resourceClient.Create(context.TODO(), baseSecretWithManagedLabel)).To(gomega.Succeed())
		requestForSecretWithManagedLabel := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: baseSecretWithManagedLabel.GetNamespace(), Name: baseSecretWithManagedLabel.GetName()}}

		baseSecret := newFixBaseSecret(testCfg.BaseNamespace, "unsuccessful-propagation")
		createSecret(g, resourceClient, baseSecret)
		defer deleteSecret(g, k8sClient, baseSecret)
		_, err := reconciler.Reconcile(ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: baseSecretWithManagedLabel.GetNamespace(),
				Name:      "not-existing-secret",
			},
		})
		g.Expect(err).To(gomega.BeNil())

		t.Log("reconciling the Secret")
		result, err := reconciler.Reconcile(requestForSecretWithManagedLabel)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.Requeue).To(gomega.BeFalse())
		g.Expect(result.RequeueAfter).To(gomega.Equal(testCfg.SecretRequeueDuration))

		updatedBase := &corev1.Secret{}
		g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: baseSecretWithManagedLabel.GetNamespace(), Name: baseSecretWithManagedLabel.GetName()}, updatedBase)).To(gomega.Succeed())
		secret := &corev1.Secret{}
		g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseSecretWithManagedLabel.GetName()}, secret)).To(gomega.Succeed())
		compareSecrets(g, secret, updatedBase)

		t.Log("updating the base Secret")
		updateBaseSecretCopy := updatedBase.DeepCopy()
		updateBaseSecretCopy.Data["test123"] = []byte("321tset")
		g.Expect(k8sClient.Update(context.TODO(), updateBaseSecretCopy)).To(gomega.Succeed())

		result, err = reconciler.Reconcile(requestForSecretWithManagedLabel)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.Requeue).To(gomega.BeFalse())
		g.Expect(result.RequeueAfter).To(gomega.Equal(testCfg.SecretRequeueDuration))

		secret = &corev1.Secret{}
		g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseSecretWithManagedLabel.GetName()}, secret)).To(gomega.Succeed())
		compareSecrets(g, secret, updatedBase)
	})

	t.Run("should successfully delete propagated Secrets from user namespace when base Secret is deleted", func(t *testing.T) {
		//GIVEN
		g := gomega.NewGomegaWithT(t)
		t.Log("reconciling the Secret")
		baseSecret := newFixBaseSecret(testCfg.BaseNamespace, "successful-deletion")
		createSecret(g, resourceClient, baseSecret)

		request := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: testCfg.BaseNamespace, Name: baseSecret.GetName()}}

		//WHEN
		result, err := reconciler.Reconcile(request)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.Requeue).To(gomega.BeFalse())
		g.Expect(result.RequeueAfter).To(gomega.Equal(testCfg.SecretRequeueDuration))

		updatedBase := &corev1.Secret{}
		g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: baseSecret.GetNamespace(), Name: baseSecret.GetName()}, updatedBase)).To(gomega.Succeed())
		g.Expect(updatedBase.Finalizers).To(gomega.ContainElement(cfgSecretFinalizerName), "created base secret should have finalizer applied")
		secret := &corev1.Secret{}
		g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseSecret.GetName()}, secret)).To(gomega.Succeed())
		compareSecrets(g, secret, baseSecret)

		t.Log("deleting base Secret")
		g.Expect(k8sClient.Delete(context.TODO(), updatedBase)).To(gomega.Succeed())
		result, err = reconciler.Reconcile(request)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.Requeue).To(gomega.BeFalse())
		g.Expect(result.RequeueAfter).To(gomega.BeZero())
		g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: updatedBase.GetNamespace(), Name: updatedBase.GetName()}, updatedBase)).To(gomega.HaveOccurred())
		g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: secret.GetNamespace(), Name: secret.GetName()}, secret)).To(gomega.HaveOccurred())
	})

	t.Run("should not successfully delete propagated Secrets from user namespace managed by user when base Secret is deleted", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		t.Log("reconciling the Secret")
		baseSecret := newFixBaseSecret(testCfg.BaseNamespace, "unsuccessful-deletion")
		createSecret(g, resourceClient, baseSecret)
		defer deleteSecret(g, k8sClient, baseSecret)

		baseSecretWithManagedLabel := newFixBaseSecretWithManagedLabel(testCfg.BaseNamespace, "secret-with-managed-label")
		g.Expect(resourceClient.Create(context.TODO(), baseSecretWithManagedLabel)).To(gomega.Succeed())
		requestForSecretWithManagedLabel := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: baseSecretWithManagedLabel.GetNamespace(), Name: baseSecretWithManagedLabel.GetName()}}

		result, err := reconciler.Reconcile(requestForSecretWithManagedLabel)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.Requeue).To(gomega.BeFalse())
		g.Expect(result.RequeueAfter).To(gomega.Equal(testCfg.SecretRequeueDuration))

		updatedBase := &corev1.Secret{}
		g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: baseSecretWithManagedLabel.GetNamespace(), Name: baseSecretWithManagedLabel.GetName()}, updatedBase)).To(gomega.Succeed())
		secret := &corev1.Secret{}
		g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseSecretWithManagedLabel.GetName()}, secret)).To(gomega.Succeed())
		compareSecrets(g, secret, updatedBase)

		t.Log("deleting base Secret")
		g.Expect(k8sClient.Delete(context.TODO(), updatedBase)).To(gomega.Succeed())
		result, err = reconciler.Reconcile(requestForSecretWithManagedLabel)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.Requeue).To(gomega.BeFalse())
		g.Expect(result.RequeueAfter).To(gomega.BeZero())
		g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: updatedBase.GetNamespace(), Name: updatedBase.GetName()}, updatedBase)).To(gomega.HaveOccurred())
		g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: secret.GetNamespace(), Name: secret.GetName()}, secret)).To(gomega.Succeed())
	})
}

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
