package kubernetes

import (
	"context"
	"testing"

	rbacv1 "k8s.io/api/rbac/v1"

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

var _ = ginkgo.Describe("RoleBinding", func() {
	var (
		reconciler      *RoleBindingReconciler
		request         ctrl.Request
		baseRoleBinding *rbacv1.RoleBinding
		namespace       string
	)

	ginkgo.BeforeEach(func() {
		userNamespace := newFixNamespace("tam")
		gomega.Expect(resourceClient.Create(context.TODO(), userNamespace)).To(gomega.Succeed())

		baseRoleBinding = newFixBaseRoleBinding(config.BaseNamespace, "ah-tak-przeciez", userNamespace.GetName())
		gomega.Expect(resourceClient.Create(context.TODO(), baseRoleBinding)).To(gomega.Succeed())

		request = ctrl.Request{NamespacedName: types.NamespacedName{Namespace: baseRoleBinding.GetNamespace(), Name: baseRoleBinding.GetName()}}
		reconciler = NewRoleBinding(k8sClient, log.Log, config, roleBindingSvc)
		namespace = userNamespace.GetName()
	})

	ginkgo.It("should successfully propagate base RoleBinding to user namespace", func() {
		ginkgo.By("reconciling RoleBinding that doesn't exist")
		_, err := reconciler.Reconcile(ctrl.Request{NamespacedName: types.NamespacedName{Namespace: baseRoleBinding.GetNamespace(), Name: "not-existing-rolebinding"}})
		gomega.Expect(err).To(gomega.BeNil(), "should not throw error on non existing RoleBinding")

		ginkgo.By("reconciling the RoleBinding")
		result, err := reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(config.RoleBindingRequeueDuration))

		roleBinding := &rbacv1.RoleBinding{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseRoleBinding.GetName()}, roleBinding)).To(gomega.Succeed())
		compareRoleBinding(roleBinding, baseRoleBinding)

		ginkgo.By("updating the base RoleBinding")
		copy := baseRoleBinding.DeepCopy()
		copy.Labels["test"] = "value"
		copy.Subjects = []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "testname",
				Namespace: namespace,
			},
		}
		gomega.Expect(k8sClient.Update(context.TODO(), copy)).To(gomega.Succeed())

		result, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(config.RoleBindingRequeueDuration))

		roleBinding = &rbacv1.RoleBinding{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseRoleBinding.GetName()}, roleBinding)).To(gomega.Succeed())
		compareRoleBinding(roleBinding, copy)

		ginkgo.By("updating the modified RoleBinding in user namespace")
		userCopy := roleBinding.DeepCopy()
		userCopy.Subjects = []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "testname2",
				Namespace: namespace,
			},
		}
		gomega.Expect(k8sClient.Update(context.TODO(), userCopy)).To(gomega.Succeed())

		result, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(config.RoleBindingRequeueDuration))

		roleBinding = &rbacv1.RoleBinding{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseRoleBinding.GetName()}, roleBinding)).To(gomega.Succeed())
		compareRoleBinding(roleBinding, copy)
	})
})

func TestRoleBindingReconciler_predicate(t *testing.T) {
	gm := gomega.NewGomegaWithT(t)
	baseNs := "base_ns"

	r := &RoleBindingReconciler{svc: NewRoleBindingService(resource.New(&automock.K8sClient{}, runtime.NewScheme()), Config{BaseNamespace: baseNs})}
	preds := r.predicate()

	correctMeta := metav1.ObjectMeta{
		Namespace: baseNs,
		Labels:    map[string]string{RbacLabel: RoleBindingLabelValue},
	}

	pod := &corev1.Pod{ObjectMeta: correctMeta}
	labelledRoleBinding := &rbacv1.RoleBinding{ObjectMeta: correctMeta}
	unlabelledRoleBinding := &rbacv1.RoleBinding{}

	t.Run("deleteFunc", func(t *testing.T) {
		deleteEventPod := event.DeleteEvent{Meta: pod.GetObjectMeta(), Object: pod}
		deleteEventLabelledSrvAcc := event.DeleteEvent{Meta: labelledRoleBinding.GetObjectMeta(), Object: labelledRoleBinding}
		deleteEventUnlabelledSrvAcc := event.DeleteEvent{Meta: unlabelledRoleBinding.GetObjectMeta(), Object: unlabelledRoleBinding}

		gm.Expect(preds.Delete(deleteEventPod)).To(gomega.BeFalse())
		gm.Expect(preds.Delete(deleteEventLabelledSrvAcc)).To(gomega.BeFalse())
		gm.Expect(preds.Delete(deleteEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		gm.Expect(preds.Delete(event.DeleteEvent{})).To(gomega.BeFalse())
	})

	t.Run("createFunc", func(t *testing.T) {
		createEventPod := event.CreateEvent{Meta: pod.GetObjectMeta(), Object: pod}
		createEventLabelledSrvAcc := event.CreateEvent{Meta: labelledRoleBinding.GetObjectMeta(), Object: labelledRoleBinding}
		createEventUnlabelledSrvAcc := event.CreateEvent{Meta: unlabelledRoleBinding.GetObjectMeta(), Object: unlabelledRoleBinding}

		gm.Expect(preds.Create(createEventPod)).To(gomega.BeFalse())
		gm.Expect(preds.Create(createEventLabelledSrvAcc)).To(gomega.BeTrue())
		gm.Expect(preds.Create(createEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		gm.Expect(preds.Create(event.CreateEvent{})).To(gomega.BeFalse())
	})

	t.Run("genericFunc", func(t *testing.T) {
		genericEventPod := event.GenericEvent{Meta: pod.GetObjectMeta(), Object: pod}
		genericEventLabelledSrvAcc := event.GenericEvent{Meta: labelledRoleBinding.GetObjectMeta(), Object: labelledRoleBinding}
		genericEventUnlabelledSrvAcc := event.GenericEvent{Meta: unlabelledRoleBinding.GetObjectMeta(), Object: unlabelledRoleBinding}

		gm.Expect(preds.Generic(genericEventPod)).To(gomega.BeFalse())
		gm.Expect(preds.Generic(genericEventLabelledSrvAcc)).To(gomega.BeTrue())
		gm.Expect(preds.Generic(genericEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		gm.Expect(preds.Generic(event.GenericEvent{})).To(gomega.BeFalse())
	})

	t.Run("updateFunc", func(t *testing.T) {
		updateEventPod := event.UpdateEvent{MetaNew: pod.GetObjectMeta(), ObjectNew: pod}
		updateEventLabelledSrvAcc := event.UpdateEvent{MetaNew: labelledRoleBinding.GetObjectMeta(), ObjectNew: labelledRoleBinding}
		updateEventUnlabelledSrvAcc := event.UpdateEvent{MetaNew: unlabelledRoleBinding.GetObjectMeta(), ObjectNew: unlabelledRoleBinding}

		gm.Expect(preds.Update(updateEventPod)).To(gomega.BeFalse())
		gm.Expect(preds.Update(updateEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		gm.Expect(preds.Update(updateEventLabelledSrvAcc)).To(gomega.BeTrue())
		gm.Expect(preds.Update(event.UpdateEvent{})).To(gomega.BeFalse())
	})
}
