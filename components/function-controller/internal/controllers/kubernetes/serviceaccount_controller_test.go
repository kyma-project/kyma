package kubernetes

import (
	"context"
	"testing"

	"k8s.io/client-go/kubernetes/scheme"

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

func TestServiceAccountReconciler_Reconcile(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	k8sClient, testEnv := setUpTestEnv(g)
	defer tearDownTestEnv(g, testEnv)
	resourceClient := resource.New(k8sClient, scheme.Scheme)
	testCfg := setUpControllerConfig(g)
	serviceAccountSvc := NewServiceAccountService(resourceClient, testCfg)

	baseNamespace := newFixNamespace(testCfg.BaseNamespace)
	g.Expect(k8sClient.Create(context.TODO(), baseNamespace)).To(gomega.Succeed())

	userNamespace := newFixNamespace("tam")
	g.Expect(resourceClient.Create(context.TODO(), userNamespace)).To(gomega.Succeed())

	baseServiceAccount := newFixBaseServiceAccount(testCfg.BaseNamespace, "ah-tak-przeciez")
	g.Expect(resourceClient.Create(context.TODO(), baseServiceAccount)).To(gomega.Succeed())

	request := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: baseServiceAccount.GetNamespace(), Name: baseServiceAccount.GetName()}}
	reconciler := NewServiceAccount(k8sClient, log.Log, testCfg, serviceAccountSvc)
	namespace := userNamespace.GetName()

	t.Run("should successfully propagate base ServiceAccount to user namespace", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)

		t.Log("reconciling the non existing Service Account")
		_, err := reconciler.Reconcile(ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: baseServiceAccount.GetNamespace(),
				Name:      "non-existing-svc-acc",
			},
		})
		g.Expect(err).To(gomega.BeNil())

		t.Log("reconciling the Service Account")
		result, err := reconciler.Reconcile(request)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.Requeue).To(gomega.BeFalse())
		g.Expect(result.RequeueAfter).To(gomega.Equal(testCfg.ServiceAccountRequeueDuration))

		serviceAccount := &corev1.ServiceAccount{}
		g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseServiceAccount.GetName()}, serviceAccount)).To(gomega.Succeed())
		compareServiceAccounts(g, serviceAccount, baseServiceAccount)

		t.Log("updating the base ServiceAccount")
		baseServiceAccountCopy := baseServiceAccount.DeepCopy()
		baseServiceAccountCopy.Labels["test"] = "value"
		baseServiceAccountCopy.AutomountServiceAccountToken = nil
		g.Expect(k8sClient.Update(context.TODO(), baseServiceAccountCopy)).To(gomega.Succeed())

		result, err = reconciler.Reconcile(request)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.Requeue).To(gomega.BeFalse())
		g.Expect(result.RequeueAfter).To(gomega.Equal(testCfg.ServiceAccountRequeueDuration))

		serviceAccount = &corev1.ServiceAccount{}
		g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseServiceAccount.GetName()}, serviceAccount)).To(gomega.Succeed())
		compareServiceAccounts(g, serviceAccount, baseServiceAccountCopy)

		t.Log("updating the modified ServiceAccount in user namespace")
		userCopy := serviceAccount.DeepCopy()
		trueValue := true
		userCopy.AutomountServiceAccountToken = &trueValue
		g.Expect(k8sClient.Update(context.TODO(), userCopy)).To(gomega.Succeed())

		result, err = reconciler.Reconcile(request)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.Requeue).To(gomega.BeFalse())
		g.Expect(result.RequeueAfter).To(gomega.Equal(testCfg.ServiceAccountRequeueDuration))

		serviceAccount = &corev1.ServiceAccount{}
		g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseServiceAccount.GetName()}, serviceAccount)).To(gomega.Succeed())
		compareServiceAccounts(g, serviceAccount, baseServiceAccountCopy)
	})
}

func TestServiceAccountReconciler_getPredicates(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	baseNs := "base_ns"

	r := &ServiceAccountReconciler{svc: NewServiceAccountService(resource.New(&automock.K8sClient{}, runtime.NewScheme()), Config{BaseNamespace: baseNs})}
	preds := r.predicate()

	correctMeta := metav1.ObjectMeta{
		Namespace: baseNs,
		Labels:    map[string]string{ConfigLabel: ServiceAccountLabelValue},
	}

	pod := &corev1.Pod{ObjectMeta: correctMeta}
	labelledSrvAcc := &corev1.ServiceAccount{ObjectMeta: correctMeta}
	unlabelledSrvAcc := &corev1.ServiceAccount{}

	t.Run("deleteFunc", func(t *testing.T) {
		deleteEventPod := event.DeleteEvent{Meta: pod.GetObjectMeta(), Object: pod}
		deleteEventLabelledSrvAcc := event.DeleteEvent{Meta: labelledSrvAcc.GetObjectMeta(), Object: labelledSrvAcc}
		deleteEventUnlabelledSrvAcc := event.DeleteEvent{Meta: unlabelledSrvAcc.GetObjectMeta(), Object: unlabelledSrvAcc}

		g.Expect(preds.Delete(deleteEventPod)).To(gomega.BeFalse())
		g.Expect(preds.Delete(deleteEventLabelledSrvAcc)).To(gomega.BeFalse())
		g.Expect(preds.Delete(deleteEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		g.Expect(preds.Delete(event.DeleteEvent{})).To(gomega.BeFalse())
	})

	t.Run("createFunc", func(t *testing.T) {
		createEventPod := event.CreateEvent{Meta: pod.GetObjectMeta(), Object: pod}
		createEventLabelledSrvAcc := event.CreateEvent{Meta: labelledSrvAcc.GetObjectMeta(), Object: labelledSrvAcc}
		createEventUnlabelledSrvAcc := event.CreateEvent{Meta: unlabelledSrvAcc.GetObjectMeta(), Object: unlabelledSrvAcc}

		g.Expect(preds.Create(createEventPod)).To(gomega.BeFalse())
		g.Expect(preds.Create(createEventLabelledSrvAcc)).To(gomega.BeTrue())
		g.Expect(preds.Create(createEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		g.Expect(preds.Create(event.CreateEvent{})).To(gomega.BeFalse())
	})

	t.Run("genericFunc", func(t *testing.T) {
		genericEventPod := event.GenericEvent{Meta: pod.GetObjectMeta(), Object: pod}
		genericEventLabelledSrvAcc := event.GenericEvent{Meta: labelledSrvAcc.GetObjectMeta(), Object: labelledSrvAcc}
		genericEventUnlabelledSrvAcc := event.GenericEvent{Meta: unlabelledSrvAcc.GetObjectMeta(), Object: unlabelledSrvAcc}

		g.Expect(preds.Generic(genericEventPod)).To(gomega.BeFalse())
		g.Expect(preds.Generic(genericEventLabelledSrvAcc)).To(gomega.BeTrue())
		g.Expect(preds.Generic(genericEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		g.Expect(preds.Generic(event.GenericEvent{})).To(gomega.BeFalse())
	})

	t.Run("updateFunc", func(t *testing.T) {
		updateEventPod := event.UpdateEvent{MetaNew: pod.GetObjectMeta(), ObjectNew: pod}
		updateEventLabelledSrvAcc := event.UpdateEvent{MetaNew: labelledSrvAcc.GetObjectMeta(), ObjectNew: labelledSrvAcc}
		updateEventUnlabelledSrvAcc := event.UpdateEvent{MetaNew: unlabelledSrvAcc.GetObjectMeta(), ObjectNew: unlabelledSrvAcc}

		g.Expect(preds.Update(updateEventPod)).To(gomega.BeFalse())
		g.Expect(preds.Update(updateEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		g.Expect(preds.Update(updateEventLabelledSrvAcc)).To(gomega.BeTrue())
		g.Expect(preds.Update(event.UpdateEvent{})).To(gomega.BeFalse())
	})
}
