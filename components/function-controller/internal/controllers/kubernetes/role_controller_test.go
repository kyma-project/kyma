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

var _ = ginkgo.Describe("Role", func() {
	var (
		reconciler *RoleReconciler
		request    ctrl.Request
		baseRole   *rbacv1.Role
		namespace  string
	)

	ginkgo.BeforeEach(func() {
		userNamespace := newFixNamespace("tam")
		gomega.Expect(resourceClient.Create(context.TODO(), userNamespace)).To(gomega.Succeed())

		baseRole = newFixBaseRole(config.BaseNamespace, "ah-tak-przeciez")
		gomega.Expect(resourceClient.Create(context.TODO(), baseRole)).To(gomega.Succeed())

		request = ctrl.Request{NamespacedName: types.NamespacedName{Namespace: baseRole.GetNamespace(), Name: baseRole.GetName()}}
		reconciler = NewRole(k8sClient, log.Log, config, roleSvc)
		namespace = userNamespace.GetName()
	})

	ginkgo.It("should successfully propagate base Role to user namespace", func() {
		ginkgo.By("reconciling Role that doesn't exist")
		_, err := reconciler.Reconcile(ctrl.Request{NamespacedName: types.NamespacedName{Namespace: baseRole.GetNamespace(), Name: "not-existing-role"}})
		gomega.Expect(err).To(gomega.BeNil(), "should not throw error on non existing Role")

		ginkgo.By("reconciling the Role")
		result, err := reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(config.RoleRequeueDuration))

		role := &rbacv1.Role{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseRole.GetName()}, role)).To(gomega.Succeed())
		compareRole(role, baseRole)

		ginkgo.By("updating the base Role")
		copy := baseRole.DeepCopy()
		copy.Labels["test"] = "value"
		copy.Rules = []rbacv1.PolicyRule{
			{
				Verbs:         []string{"get"},
				Resources:     []string{"configmaps"},
				ResourceNames: []string{"test"},
			},
		}
		gomega.Expect(k8sClient.Update(context.TODO(), copy)).To(gomega.Succeed())

		result, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(config.RoleRequeueDuration))

		role = &rbacv1.Role{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseRole.GetName()}, role)).To(gomega.Succeed())
		compareRole(role, copy)

		ginkgo.By("updating the modified Role in user namespace")
		userCopy := role.DeepCopy()
		userCopy.Rules = []rbacv1.PolicyRule{
			{
				Verbs:         []string{"create"},
				Resources:     []string{"secrets"},
				ResourceNames: []string{"test2"},
			},
		}
		gomega.Expect(k8sClient.Update(context.TODO(), userCopy)).To(gomega.Succeed())

		result, err = reconciler.Reconcile(request)
		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(result.Requeue).To(gomega.BeFalse())
		gomega.Expect(result.RequeueAfter).To(gomega.Equal(config.RoleRequeueDuration))

		role = &rbacv1.Role{}
		gomega.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseRole.GetName()}, role)).To(gomega.Succeed())
		compareRole(role, copy)
	})
})

func TestRoleReconciler_predicate(t *testing.T) {
	gm := gomega.NewGomegaWithT(t)
	baseNs := "base_ns"

	r := &RoleReconciler{svc: NewRoleService(resource.New(&automock.K8sClient{}, runtime.NewScheme()), Config{BaseNamespace: baseNs})}
	preds := r.predicate()

	correctMeta := metav1.ObjectMeta{
		Namespace: baseNs,
		Labels:    map[string]string{RbacLabel: RoleLabelValue},
	}

	pod := &corev1.Pod{ObjectMeta: correctMeta}
	labelledRole := &rbacv1.Role{ObjectMeta: correctMeta}
	unlabelledRole := &rbacv1.Role{}

	t.Run("deleteFunc", func(t *testing.T) {
		deleteEventPod := event.DeleteEvent{Meta: pod.GetObjectMeta(), Object: pod}
		deleteEventLabelledSrvAcc := event.DeleteEvent{Meta: labelledRole.GetObjectMeta(), Object: labelledRole}
		deleteEventUnlabelledSrvAcc := event.DeleteEvent{Meta: unlabelledRole.GetObjectMeta(), Object: unlabelledRole}

		gm.Expect(preds.Delete(deleteEventPod)).To(gomega.BeFalse())
		gm.Expect(preds.Delete(deleteEventLabelledSrvAcc)).To(gomega.BeFalse())
		gm.Expect(preds.Delete(deleteEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		gm.Expect(preds.Delete(event.DeleteEvent{})).To(gomega.BeFalse())
	})

	t.Run("createFunc", func(t *testing.T) {
		createEventPod := event.CreateEvent{Meta: pod.GetObjectMeta(), Object: pod}
		createEventLabelledSrvAcc := event.CreateEvent{Meta: labelledRole.GetObjectMeta(), Object: labelledRole}
		createEventUnlabelledSrvAcc := event.CreateEvent{Meta: unlabelledRole.GetObjectMeta(), Object: unlabelledRole}

		gm.Expect(preds.Create(createEventPod)).To(gomega.BeFalse())
		gm.Expect(preds.Create(createEventLabelledSrvAcc)).To(gomega.BeTrue())
		gm.Expect(preds.Create(createEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		gm.Expect(preds.Create(event.CreateEvent{})).To(gomega.BeFalse())
	})

	t.Run("genericFunc", func(t *testing.T) {
		genericEventPod := event.GenericEvent{Meta: pod.GetObjectMeta(), Object: pod}
		genericEventLabelledSrvAcc := event.GenericEvent{Meta: labelledRole.GetObjectMeta(), Object: labelledRole}
		genericEventUnlabelledSrvAcc := event.GenericEvent{Meta: unlabelledRole.GetObjectMeta(), Object: unlabelledRole}

		gm.Expect(preds.Generic(genericEventPod)).To(gomega.BeFalse())
		gm.Expect(preds.Generic(genericEventLabelledSrvAcc)).To(gomega.BeTrue())
		gm.Expect(preds.Generic(genericEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		gm.Expect(preds.Generic(event.GenericEvent{})).To(gomega.BeFalse())
	})

	t.Run("updateFunc", func(t *testing.T) {
		updateEventPod := event.UpdateEvent{MetaNew: pod.GetObjectMeta(), ObjectNew: pod}
		updateEventLabelledSrvAcc := event.UpdateEvent{MetaNew: labelledRole.GetObjectMeta(), ObjectNew: labelledRole}
		updateEventUnlabelledSrvAcc := event.UpdateEvent{MetaNew: unlabelledRole.GetObjectMeta(), ObjectNew: unlabelledRole}

		gm.Expect(preds.Update(updateEventPod)).To(gomega.BeFalse())
		gm.Expect(preds.Update(updateEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		gm.Expect(preds.Update(updateEventLabelledSrvAcc)).To(gomega.BeTrue())
		gm.Expect(preds.Update(event.UpdateEvent{})).To(gomega.BeFalse())
	})
}
