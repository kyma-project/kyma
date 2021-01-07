package kubernetes

import (
	"context"
	"testing"
	"time"

	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var _ = ginkgo.Describe("Namespace", func() {
	var (
		reconciler         *NamespaceReconciler
		request            ctrl.Request
		baseSecret         *corev1.Secret
		baseConfigMap      *corev1.ConfigMap
		baseServiceAccount *corev1.ServiceAccount
		baseRole           *rbacv1.Role
		baseRoleBinding    *rbacv1.RoleBinding
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

		baseRole = newFixBaseRole(config.BaseNamespace, "ah-tak-przeciez")
		gomega.Expect(resourceClient.Create(context.TODO(), baseRole)).To(gomega.Succeed())

		baseRoleBinding = newFixBaseRoleBinding(config.BaseNamespace, "ah-tak-przeciez")
		gomega.Expect(resourceClient.Create(context.TODO(), baseRoleBinding)).To(gomega.Succeed())

		request = ctrl.Request{NamespacedName: types.NamespacedName{Name: userNamespace.GetName()}}
		reconciler = NewNamespace(k8sClient, log.Log, config, configMapSvc, secretSvc, serviceAccountSvc, roleSvc, roleBindingSvc)
		namespace = userNamespace.GetName()
	})

	ginkgo.It("should successfully propagate base ServiceAccount to user namespace", func() {
		ginkgo.By("reconciling Namespace that doesn't exist")
		_, err := reconciler.Reconcile(ctrl.Request{NamespacedName: types.NamespacedName{Name: "not-existing-ns"}})
		gomega.Expect(err).To(gomega.BeNil(), "should not throw error on non existing namespace")

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

		role := &rbacv1.Role{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseRole.GetName()}, role)).To(gomega.Succeed())
		compareRole(role, baseRole)

		roleBinding := &rbacv1.RoleBinding{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseRoleBinding.GetName()}, roleBinding)).To(gomega.Succeed())
		compareRoleBinding(roleBinding, baseRoleBinding)

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

		role = &rbacv1.Role{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseRole.GetName()}, role)).To(gomega.Succeed())
		compareRole(role, baseRole)

		roleBinding = &rbacv1.RoleBinding{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseRoleBinding.GetName()}, roleBinding)).To(gomega.Succeed())
		compareRoleBinding(roleBinding, baseRoleBinding)
	})
})

func TestNamespaceReconciler_predicate(t *testing.T) {
	gm := gomega.NewGomegaWithT(t)

	baseNs := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "base-ns"}}
	excludedNs := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "excluded-1"}}
	normalNs := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "normal-1"}}
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod-name"}}

	r := &NamespaceReconciler{config: Config{
		BaseNamespace:      baseNs.Name,
		ExcludedNamespaces: []string{excludedNs.Name},
	}}
	preds := r.predicate()

	t.Run("deleteFunc", func(t *testing.T) {
		podEvent := event.DeleteEvent{Meta: pod.GetObjectMeta(), Object: pod}
		eventBaseNs := event.DeleteEvent{Meta: baseNs.GetObjectMeta(), Object: baseNs}
		eventExcludedNs := event.DeleteEvent{Meta: excludedNs.GetObjectMeta(), Object: excludedNs}
		normalNsEvent := event.DeleteEvent{Meta: normalNs.GetObjectMeta(), Object: normalNs}

		gm.Expect(preds.Delete(podEvent)).To(gomega.BeFalse())
		gm.Expect(preds.Delete(eventBaseNs)).To(gomega.BeFalse())
		gm.Expect(preds.Delete(eventExcludedNs)).To(gomega.BeFalse())
		gm.Expect(preds.Delete(normalNsEvent)).To(gomega.BeFalse())
		gm.Expect(preds.Delete(event.DeleteEvent{})).To(gomega.BeFalse())
	})

	t.Run("createFunc", func(t *testing.T) {
		podEvent := event.CreateEvent{Meta: pod.GetObjectMeta(), Object: pod}
		eventBaseNs := event.CreateEvent{Meta: baseNs.GetObjectMeta(), Object: baseNs}
		eventExcludedNs := event.CreateEvent{Meta: excludedNs.GetObjectMeta(), Object: excludedNs}
		normalNsEvent := event.CreateEvent{Meta: normalNs.GetObjectMeta(), Object: normalNs}

		gm.Expect(preds.Create(podEvent)).To(gomega.BeFalse())
		gm.Expect(preds.Create(eventBaseNs)).To(gomega.BeFalse())
		gm.Expect(preds.Create(eventExcludedNs)).To(gomega.BeFalse())
		gm.Expect(preds.Create(event.CreateEvent{})).To(gomega.BeFalse())

		gm.Expect(preds.Create(normalNsEvent)).To(gomega.BeTrue(), "should be true for non-base, non-excluded ns")
	})

	t.Run("genericFunc", func(t *testing.T) {
		podEvent := event.GenericEvent{Meta: pod.GetObjectMeta(), Object: pod}
		eventBaseNs := event.GenericEvent{Meta: baseNs.GetObjectMeta(), Object: baseNs}
		eventExcludedNs := event.GenericEvent{Meta: excludedNs.GetObjectMeta(), Object: excludedNs}
		normalNsEvent := event.GenericEvent{Meta: normalNs.GetObjectMeta(), Object: normalNs}

		gm.Expect(preds.Generic(podEvent)).To(gomega.BeFalse())
		gm.Expect(preds.Generic(eventBaseNs)).To(gomega.BeFalse())
		gm.Expect(preds.Generic(eventExcludedNs)).To(gomega.BeFalse())
		gm.Expect(preds.Generic(event.GenericEvent{})).To(gomega.BeFalse())
		gm.Expect(preds.Generic(normalNsEvent)).To(gomega.BeFalse())
	})

	t.Run("updateFunc", func(t *testing.T) {
		podEvent := event.UpdateEvent{MetaNew: pod.GetObjectMeta(), ObjectNew: pod}
		eventBaseNs := event.UpdateEvent{MetaNew: baseNs.GetObjectMeta(), ObjectNew: baseNs}
		eventExcludedNs := event.UpdateEvent{MetaNew: excludedNs.GetObjectMeta(), ObjectNew: excludedNs}
		normalNsEvent := event.UpdateEvent{MetaNew: normalNs.GetObjectMeta(), ObjectNew: normalNs}

		gm.Expect(preds.Update(podEvent)).To(gomega.BeFalse())
		gm.Expect(preds.Update(eventBaseNs)).To(gomega.BeFalse())
		gm.Expect(preds.Update(eventExcludedNs)).To(gomega.BeFalse())
		gm.Expect(preds.Update(event.UpdateEvent{})).To(gomega.BeFalse())
		gm.Expect(preds.Update(normalNsEvent)).To(gomega.BeFalse())
	})
}
