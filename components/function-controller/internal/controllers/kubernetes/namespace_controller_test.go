package kubernetes

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
	"k8s.io/client-go/kubernetes/scheme"

	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func TestNamespaceReconciler_Reconcile(t *testing.T) {
	//GIVEN
	g := gomega.NewGomegaWithT(t)
	k8sClient, testEnv := setUpTestEnv(g)
	defer tearDownTestEnv(g, testEnv)
	resourceClient := resource.New(k8sClient, scheme.Scheme)
	testCfg := setUpControllerConfig(g)

	configMapSvc := NewConfigMapService(resourceClient, testCfg)
	secretSvc := NewSecretService(resourceClient, testCfg)
	serviceAccountSvc := NewServiceAccountService(resourceClient, testCfg)
	roleSvc := NewRoleService(resourceClient, testCfg)
	roleBindingSvc := NewRoleBindingService(resourceClient, testCfg)

	cfgNamespace := newFixNamespace(testCfg.BaseNamespace)
	g.Expect(k8sClient.Create(context.TODO(), cfgNamespace)).To(gomega.Succeed())

	userNamespace := newFixNamespace("tam")
	g.Expect(k8sClient.Create(context.TODO(), userNamespace)).To(gomega.Succeed())

	baseConfigMap := newFixBaseConfigMap(testCfg.BaseNamespace, "ah-tak-przeciez")
	g.Expect(k8sClient.Create(context.TODO(), baseConfigMap)).To(gomega.Succeed())

	baseSecret := newFixBaseSecret(testCfg.BaseNamespace, "ah-tak-przeciez")
	g.Expect(k8sClient.Create(context.TODO(), baseSecret)).To(gomega.Succeed())

	baseServiceAccount := newFixBaseServiceAccount(testCfg.BaseNamespace, "ah-tak-przeciez")
	g.Expect(k8sClient.Create(context.TODO(), baseServiceAccount)).To(gomega.Succeed())

	baseRole := newFixBaseRole(testCfg.BaseNamespace, "ah-tak-przeciez")
	g.Expect(k8sClient.Create(context.TODO(), baseRole)).To(gomega.Succeed())

	baseRoleBinding := newFixBaseRoleBinding(testCfg.BaseNamespace, "ah-tak-przeciez", userNamespace.GetName())
	g.Expect(k8sClient.Create(context.TODO(), baseRoleBinding)).To(gomega.Succeed())

	request := ctrl.Request{NamespacedName: types.NamespacedName{Name: userNamespace.GetName()}}
	reconciler := NewNamespace(k8sClient, log.Log, testCfg, configMapSvc, secretSvc, serviceAccountSvc, roleSvc, roleBindingSvc)
	namespace := userNamespace.GetName()

	//WHEN
	t.Log("reconciling Namespace that doesn't exist")
	_, err := reconciler.Reconcile(ctrl.Request{NamespacedName: types.NamespacedName{Name: "not-existing-ns"}})
	g.Expect(err).To(gomega.BeNil(), "should not throw error on non existing namespace")

	t.Log("reconciling the Namespace")
	result, err := reconciler.Reconcile(request)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(result.Requeue).To(gomega.BeFalse())
	g.Expect(result.RequeueAfter).To(gomega.Equal(0 * time.Second))

	configMap := &corev1.ConfigMap{}
	g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseConfigMap.GetName()}, configMap)).To(gomega.Succeed())
	compareConfigMaps(g, configMap, baseConfigMap)

	secret := &corev1.Secret{}
	g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseSecret.GetName()}, secret)).To(gomega.Succeed())
	compareSecrets(g, secret, baseSecret)

	serviceAccount := &corev1.ServiceAccount{}
	g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseServiceAccount.GetName()}, serviceAccount)).To(gomega.Succeed())
	compareServiceAccounts(g, serviceAccount, baseServiceAccount)

	role := &rbacv1.Role{}
	g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseRole.GetName()}, role)).To(gomega.Succeed())
	compareRole(g, role, baseRole)

	roleBinding := &rbacv1.RoleBinding{}
	g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseRoleBinding.GetName()}, roleBinding)).To(gomega.Succeed())
	compareRoleBinding(g, roleBinding, baseRoleBinding)

	t.Log("one more time reconciling the Namespace")
	result, err = reconciler.Reconcile(request)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(result.Requeue).To(gomega.BeFalse())
	g.Expect(result.RequeueAfter).To(gomega.Equal(0 * time.Second))

	configMap = &corev1.ConfigMap{}
	g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseConfigMap.GetName()}, configMap)).To(gomega.Succeed())
	compareConfigMaps(g, configMap, baseConfigMap)

	secret = &corev1.Secret{}
	g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseSecret.GetName()}, secret)).To(gomega.Succeed())
	compareSecrets(g, secret, baseSecret)

	serviceAccount = &corev1.ServiceAccount{}
	g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseServiceAccount.GetName()}, serviceAccount)).To(gomega.Succeed())
	compareServiceAccounts(g, serviceAccount, baseServiceAccount)

	role = &rbacv1.Role{}
	g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseRole.GetName()}, role)).To(gomega.Succeed())
	compareRole(g, role, baseRole)

	roleBinding = &rbacv1.RoleBinding{}
	g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseRoleBinding.GetName()}, roleBinding)).To(gomega.Succeed())
	compareRoleBinding(g, roleBinding, baseRoleBinding)
}

func TestNamespaceReconciler_predicate(t *testing.T) {
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
		g := gomega.NewGomegaWithT(t)
		podEvent := event.DeleteEvent{Meta: pod.GetObjectMeta(), Object: pod}
		eventBaseNs := event.DeleteEvent{Meta: baseNs.GetObjectMeta(), Object: baseNs}
		eventExcludedNs := event.DeleteEvent{Meta: excludedNs.GetObjectMeta(), Object: excludedNs}
		normalNsEvent := event.DeleteEvent{Meta: normalNs.GetObjectMeta(), Object: normalNs}

		g.Expect(preds.Delete(podEvent)).To(gomega.BeFalse())
		g.Expect(preds.Delete(eventBaseNs)).To(gomega.BeFalse())
		g.Expect(preds.Delete(eventExcludedNs)).To(gomega.BeFalse())
		g.Expect(preds.Delete(normalNsEvent)).To(gomega.BeFalse())
		g.Expect(preds.Delete(event.DeleteEvent{})).To(gomega.BeFalse())
	})

	t.Run("createFunc", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		podEvent := event.CreateEvent{Meta: pod.GetObjectMeta(), Object: pod}
		eventBaseNs := event.CreateEvent{Meta: baseNs.GetObjectMeta(), Object: baseNs}
		eventExcludedNs := event.CreateEvent{Meta: excludedNs.GetObjectMeta(), Object: excludedNs}
		normalNsEvent := event.CreateEvent{Meta: normalNs.GetObjectMeta(), Object: normalNs}

		g.Expect(preds.Create(podEvent)).To(gomega.BeFalse())
		g.Expect(preds.Create(eventBaseNs)).To(gomega.BeFalse())
		g.Expect(preds.Create(eventExcludedNs)).To(gomega.BeFalse())
		g.Expect(preds.Create(event.CreateEvent{})).To(gomega.BeFalse())

		g.Expect(preds.Create(normalNsEvent)).To(gomega.BeTrue(), "should be true for non-base, non-excluded ns")
	})

	t.Run("genericFunc", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		podEvent := event.GenericEvent{Meta: pod.GetObjectMeta(), Object: pod}
		eventBaseNs := event.GenericEvent{Meta: baseNs.GetObjectMeta(), Object: baseNs}
		eventExcludedNs := event.GenericEvent{Meta: excludedNs.GetObjectMeta(), Object: excludedNs}
		normalNsEvent := event.GenericEvent{Meta: normalNs.GetObjectMeta(), Object: normalNs}

		g.Expect(preds.Generic(podEvent)).To(gomega.BeFalse())
		g.Expect(preds.Generic(eventBaseNs)).To(gomega.BeFalse())
		g.Expect(preds.Generic(eventExcludedNs)).To(gomega.BeFalse())
		g.Expect(preds.Generic(event.GenericEvent{})).To(gomega.BeFalse())
		g.Expect(preds.Generic(normalNsEvent)).To(gomega.BeFalse())
	})

	t.Run("updateFunc", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		podEvent := event.UpdateEvent{MetaNew: pod.GetObjectMeta(), ObjectNew: pod}
		eventBaseNs := event.UpdateEvent{MetaNew: baseNs.GetObjectMeta(), ObjectNew: baseNs}
		eventExcludedNs := event.UpdateEvent{MetaNew: excludedNs.GetObjectMeta(), ObjectNew: excludedNs}
		normalNsEvent := event.UpdateEvent{MetaNew: normalNs.GetObjectMeta(), ObjectNew: normalNs}

		g.Expect(preds.Update(podEvent)).To(gomega.BeFalse())
		g.Expect(preds.Update(eventBaseNs)).To(gomega.BeFalse())
		g.Expect(preds.Update(eventExcludedNs)).To(gomega.BeFalse())
		g.Expect(preds.Update(event.UpdateEvent{})).To(gomega.BeFalse())
		g.Expect(preds.Update(normalNsEvent)).To(gomega.BeFalse())
	})
}
