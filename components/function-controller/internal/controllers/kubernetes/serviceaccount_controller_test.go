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

var _ = ginkgo.Describe("ServiceAccount", func() {
	var (
		reconciler         *ServiceAccountReconciler
		request            ctrl.Request
		baseServiceAccount *corev1.ServiceAccount
		namespace          string
	)
	g := gomega.NewGomegaWithT(nil)

	ginkgo.BeforeEach(func() {
		userNamespace := newFixNamespace("tam")
		gomega.Expect(resourceClient.Create(context.TODO(), userNamespace)).To(gomega.Succeed())

		baseServiceAccount = newFixBaseServiceAccount(config.BaseNamespace, "ah-tak-przeciez")
		gomega.Expect(resourceClient.Create(context.TODO(), baseServiceAccount)).To(gomega.Succeed())

		request = ctrl.Request{NamespacedName: types.NamespacedName{Namespace: baseServiceAccount.GetNamespace(), Name: baseServiceAccount.GetName()}}
		reconciler = NewServiceAccount(k8sClient, log.Log, config, serviceAccountSvc)
		namespace = userNamespace.GetName()
	})

	ginkgo.It("should successfully propagate base ServiceAccount to user namespace", func() {
		ginkgo.By("reconciling the non existing Service Account")
		_, err := reconciler.Reconcile(ctrl.Request{
			NamespacedName: types.NamespacedName{
				Namespace: baseServiceAccount.GetNamespace(),
				Name:      "non-existing-svc-acc",
			},
		})
		gomega.Expect(err).To(gomega.BeNil())

		ginkgo.By("reconciling the Service Account")
		result, err := reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(config.ServiceAccountRequeueDuration))

		serviceAccount := &corev1.ServiceAccount{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseServiceAccount.GetName()}, serviceAccount)).To(gomega.Succeed())
		compareServiceAccounts(g, serviceAccount, baseServiceAccount)

		ginkgo.By("updating the base ServiceAccount")
		copy := baseServiceAccount.DeepCopy()
		copy.Labels["test"] = "value"
		copy.AutomountServiceAccountToken = nil
		gomega.Expect(k8sClient.Update(context.TODO(), copy)).To(gomega.Succeed())

		result, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(config.ServiceAccountRequeueDuration))

		serviceAccount = &corev1.ServiceAccount{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseServiceAccount.GetName()}, serviceAccount)).To(gomega.Succeed())
		compareServiceAccounts(g, serviceAccount, copy)

		ginkgo.By("updating the modified ServiceAccount in user namespace")
		userCopy := serviceAccount.DeepCopy()
		trueValue := true
		userCopy.AutomountServiceAccountToken = &trueValue
		gomega.Expect(k8sClient.Update(context.TODO(), userCopy)).To(gomega.Succeed())

		result, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(config.ServiceAccountRequeueDuration))

		serviceAccount = &corev1.ServiceAccount{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseServiceAccount.GetName()}, serviceAccount)).To(gomega.Succeed())
		compareServiceAccounts(g, serviceAccount, copy)
	})
})

func TestServiceAccountReconciler_getPredicates(t *testing.T) {
	gm := gomega.NewGomegaWithT(t)
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

		gm.Expect(preds.Delete(deleteEventPod)).To(gomega.BeFalse())
		gm.Expect(preds.Delete(deleteEventLabelledSrvAcc)).To(gomega.BeFalse())
		gm.Expect(preds.Delete(deleteEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		gm.Expect(preds.Delete(event.DeleteEvent{})).To(gomega.BeFalse())
	})

	t.Run("createFunc", func(t *testing.T) {
		createEventPod := event.CreateEvent{Meta: pod.GetObjectMeta(), Object: pod}
		createEventLabelledSrvAcc := event.CreateEvent{Meta: labelledSrvAcc.GetObjectMeta(), Object: labelledSrvAcc}
		createEventUnlabelledSrvAcc := event.CreateEvent{Meta: unlabelledSrvAcc.GetObjectMeta(), Object: unlabelledSrvAcc}

		gm.Expect(preds.Create(createEventPod)).To(gomega.BeFalse())
		gm.Expect(preds.Create(createEventLabelledSrvAcc)).To(gomega.BeTrue())
		gm.Expect(preds.Create(createEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		gm.Expect(preds.Create(event.CreateEvent{})).To(gomega.BeFalse())
	})

	t.Run("genericFunc", func(t *testing.T) {
		genericEventPod := event.GenericEvent{Meta: pod.GetObjectMeta(), Object: pod}
		genericEventLabelledSrvAcc := event.GenericEvent{Meta: labelledSrvAcc.GetObjectMeta(), Object: labelledSrvAcc}
		genericEventUnlabelledSrvAcc := event.GenericEvent{Meta: unlabelledSrvAcc.GetObjectMeta(), Object: unlabelledSrvAcc}

		gm.Expect(preds.Generic(genericEventPod)).To(gomega.BeFalse())
		gm.Expect(preds.Generic(genericEventLabelledSrvAcc)).To(gomega.BeTrue())
		gm.Expect(preds.Generic(genericEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		gm.Expect(preds.Generic(event.GenericEvent{})).To(gomega.BeFalse())
	})

	t.Run("updateFunc", func(t *testing.T) {
		updateEventPod := event.UpdateEvent{MetaNew: pod.GetObjectMeta(), ObjectNew: pod}
		updateEventLabelledSrvAcc := event.UpdateEvent{MetaNew: labelledSrvAcc.GetObjectMeta(), ObjectNew: labelledSrvAcc}
		updateEventUnlabelledSrvAcc := event.UpdateEvent{MetaNew: unlabelledSrvAcc.GetObjectMeta(), ObjectNew: unlabelledSrvAcc}

		gm.Expect(preds.Update(updateEventPod)).To(gomega.BeFalse())
		gm.Expect(preds.Update(updateEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		gm.Expect(preds.Update(updateEventLabelledSrvAcc)).To(gomega.BeTrue())
		gm.Expect(preds.Update(event.UpdateEvent{})).To(gomega.BeFalse())
	})
}
