package kubernetes

import (
	"context"
	"testing"

	"k8s.io/client-go/kubernetes/scheme"

	rbacv1 "k8s.io/api/rbac/v1"

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

func TestRoleReconciler_Reconcile(t *testing.T) {
	//GIVEN
	g := gomega.NewGomegaWithT(t)
	k8sClient, testEnv := setUpTestEnv(g)
	defer tearDownTestEnv(g, testEnv)
	resourceClient := resource.New(k8sClient, scheme.Scheme)
	testCfg := setUpControllerConfig(g)

	baseNamespace := newFixNamespace(testCfg.BaseNamespace)
	g.Expect(k8sClient.Create(context.TODO(), baseNamespace)).To(gomega.Succeed())

	userNamespace := newFixNamespace("tam")
	g.Expect(k8sClient.Create(context.TODO(), userNamespace)).To(gomega.Succeed())

	baseRole := newFixBaseRole(testCfg.BaseNamespace, "ah-tak-przeciez")
	g.Expect(k8sClient.Create(context.TODO(), baseRole)).To(gomega.Succeed())

	roleSvc := NewRoleService(resourceClient, testCfg)
	reconciler := NewRole(k8sClient, log.Log, testCfg, roleSvc)
	namespace := userNamespace.GetName()

	request := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: baseRole.GetNamespace(), Name: baseRole.GetName()}}

	//WHEN
	t.Run("should successfully propagate base Role to user namespace", func(t *testing.T) {
		t.Log("reconciling Role that doesn't exist")
		_, err := reconciler.Reconcile(ctrl.Request{NamespacedName: types.NamespacedName{Namespace: baseRole.GetNamespace(), Name: "not-existing-role"}})
		g.Expect(err).To(gomega.BeNil(), "should not throw error on non existing Role")

		t.Log("reconciling the Role")
		result, err := reconciler.Reconcile(request)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.Requeue).To(gomega.BeFalse())
		g.Expect(result.RequeueAfter).To(gomega.Equal(testCfg.RoleRequeueDuration))

		role := &rbacv1.Role{}
		g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseRole.GetName()}, role)).To(gomega.Succeed())
		compareRole(g, role, baseRole)

		t.Log("updating the base Role")
		baseRoleCopy := baseRole.DeepCopy()
		baseRoleCopy.Labels["test"] = "value"
		baseRoleCopy.Rules = []rbacv1.PolicyRule{
			{
				APIGroups:     []string{"v1"},
				Verbs:         []string{"get"},
				Resources:     []string{"configmaps"},
				ResourceNames: []string{"test"},
			},
		}
		g.Expect(k8sClient.Update(context.TODO(), baseRoleCopy)).To(gomega.Succeed())

		result, err = reconciler.Reconcile(request)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.Requeue).To(gomega.BeFalse())
		g.Expect(result.RequeueAfter).To(gomega.Equal(testCfg.RoleRequeueDuration))

		role = &rbacv1.Role{}
		g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseRole.GetName()}, role)).To(gomega.Succeed())
		compareRole(g, role, baseRoleCopy)

		t.Log("updating the modified Role in user namespace")
		userCopy := role.DeepCopy()
		userCopy.Rules = []rbacv1.PolicyRule{
			{
				APIGroups:     []string{"v1"},
				Verbs:         []string{"create"},
				Resources:     []string{"secrets"},
				ResourceNames: []string{"test2"},
			},
		}
		g.Expect(k8sClient.Update(context.TODO(), userCopy)).To(gomega.Succeed())

		result, err = reconciler.Reconcile(request)
		g.Expect(err).To(gomega.BeNil())
		g.Expect(result.Requeue).To(gomega.BeFalse())
		g.Expect(result.RequeueAfter).To(gomega.Equal(testCfg.RoleRequeueDuration))

		role = &rbacv1.Role{}
		g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseRole.GetName()}, role)).To(gomega.Succeed())
		compareRole(g, role, baseRoleCopy)
	})
}

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
