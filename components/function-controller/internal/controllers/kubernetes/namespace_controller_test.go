package kubernetes

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
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

	request := ctrl.Request{NamespacedName: types.NamespacedName{Name: userNamespace.GetName()}}
	reconciler := NewNamespace(k8sClient, zap.NewNop().Sugar(), testCfg, configMapSvc, secretSvc, serviceAccountSvc)
	namespace := userNamespace.GetName()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//WHEN
	t.Log("reconciling Namespace that doesn't exist")
	_, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: "not-existing-ns"}})
	g.Expect(err).To(gomega.BeNil(), "should not throw error on non existing namespace")

	t.Log("reconciling the Namespace")
	result, err := reconciler.Reconcile(ctx, request)
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

	t.Log("one more time reconciling the Namespace")
	result, err = reconciler.Reconcile(ctx, request)
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
		podEvent := event.DeleteEvent{Object: pod}
		eventBaseNs := event.DeleteEvent{Object: baseNs}
		eventExcludedNs := event.DeleteEvent{Object: excludedNs}
		normalNsEvent := event.DeleteEvent{Object: normalNs}

		g.Expect(preds.Delete(podEvent)).To(gomega.BeFalse())
		g.Expect(preds.Delete(eventBaseNs)).To(gomega.BeFalse())
		g.Expect(preds.Delete(eventExcludedNs)).To(gomega.BeFalse())
		g.Expect(preds.Delete(normalNsEvent)).To(gomega.BeFalse())
		g.Expect(preds.Delete(event.DeleteEvent{})).To(gomega.BeFalse())
	})

	t.Run("createFunc", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		podEvent := event.CreateEvent{Object: pod}
		eventBaseNs := event.CreateEvent{Object: baseNs}
		eventExcludedNs := event.CreateEvent{Object: excludedNs}
		normalNsEvent := event.CreateEvent{Object: normalNs}

		g.Expect(preds.Create(podEvent)).To(gomega.BeFalse())
		g.Expect(preds.Create(eventBaseNs)).To(gomega.BeFalse())
		g.Expect(preds.Create(eventExcludedNs)).To(gomega.BeFalse())
		g.Expect(preds.Create(event.CreateEvent{})).To(gomega.BeFalse())

		g.Expect(preds.Create(normalNsEvent)).To(gomega.BeTrue(), "should be true for non-base, non-excluded ns")
	})

	t.Run("genericFunc", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		podEvent := event.GenericEvent{Object: pod}
		eventBaseNs := event.GenericEvent{Object: baseNs}
		eventExcludedNs := event.GenericEvent{Object: excludedNs}
		normalNsEvent := event.GenericEvent{Object: normalNs}

		g.Expect(preds.Generic(podEvent)).To(gomega.BeFalse())
		g.Expect(preds.Generic(eventBaseNs)).To(gomega.BeFalse())
		g.Expect(preds.Generic(eventExcludedNs)).To(gomega.BeFalse())
		g.Expect(preds.Generic(event.GenericEvent{})).To(gomega.BeFalse())
		g.Expect(preds.Generic(normalNsEvent)).To(gomega.BeFalse())
	})

	t.Run("updateFunc", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		podEvent := event.UpdateEvent{ObjectNew: pod}
		eventBaseNs := event.UpdateEvent{ObjectNew: baseNs}
		eventExcludedNs := event.UpdateEvent{ObjectNew: excludedNs}
		normalNsEvent := event.UpdateEvent{ObjectNew: normalNs}

		g.Expect(preds.Update(podEvent)).To(gomega.BeFalse())
		g.Expect(preds.Update(eventBaseNs)).To(gomega.BeFalse())
		g.Expect(preds.Update(eventExcludedNs)).To(gomega.BeFalse())
		g.Expect(preds.Update(event.UpdateEvent{})).To(gomega.BeFalse())
		g.Expect(preds.Update(normalNsEvent)).To(gomega.BeFalse())
	})
}
