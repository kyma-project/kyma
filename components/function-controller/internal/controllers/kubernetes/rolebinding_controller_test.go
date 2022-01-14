package kubernetes

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
	"github.com/kyma-project/kyma/components/function-controller/internal/resource/automock"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestRoleBindingReconciler_Reconcile(t *testing.T) {
	//GIVEN
	g := gomega.NewGomegaWithT(t)
	k8sClient, testEnv := setUpTestEnv(g)
	defer tearDownTestEnv(g, testEnv)
	resourceClient := resource.New(k8sClient, scheme.Scheme)
	testCfg := setUpControllerConfig(g)
	roleBindingSvc := NewRoleBindingService(resourceClient, testCfg)

	baseNamespace := newFixNamespace(testCfg.BaseNamespace)
	g.Expect(k8sClient.Create(context.TODO(), baseNamespace)).To(gomega.Succeed())

	userNamespace := newFixNamespace("tam")
	g.Expect(resourceClient.Create(context.TODO(), userNamespace)).To(gomega.Succeed())

	baseRoleBinding := newFixBaseRoleBinding(testCfg.BaseNamespace, "ah-tak-przeciez", userNamespace.GetName())
	g.Expect(resourceClient.Create(context.TODO(), baseRoleBinding)).To(gomega.Succeed())

	request := ctrl.Request{NamespacedName: types.NamespacedName{Namespace: baseRoleBinding.GetNamespace(), Name: baseRoleBinding.GetName()}}
	reconciler := NewRoleBinding(k8sClient, log.Log, testCfg, roleBindingSvc)
	namespace := userNamespace.GetName()

	//WHEN
	t.Log("reconciling RoleBinding that doesn't exist")
	_, err := reconciler.Reconcile(ctrl.Request{NamespacedName: types.NamespacedName{Namespace: baseRoleBinding.GetNamespace(), Name: "not-existing-rolebinding"}})
	g.Expect(err).To(gomega.BeNil(), "should not throw error on non existing RoleBinding")

	t.Log("reconciling the RoleBinding")
	result, err := reconciler.Reconcile(request)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(result.Requeue).To(gomega.BeFalse())
	g.Expect(result.RequeueAfter).To(gomega.Equal(testCfg.RoleBindingRequeueDuration))

	roleBinding := &rbacv1.RoleBinding{}
	g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseRoleBinding.GetName()}, roleBinding)).To(gomega.Succeed())
	compareRoleBinding(g, roleBinding, baseRoleBinding)

	t.Log("updating the base RoleBinding")
	roleBindingCopy := baseRoleBinding.DeepCopy()
	roleBindingCopy.Labels["test"] = "value"
	roleBindingCopy.Subjects = []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      "testname",
			Namespace: namespace,
		},
	}
	g.Expect(k8sClient.Update(context.TODO(), roleBindingCopy)).To(gomega.Succeed())

	result, err = reconciler.Reconcile(request)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(result.Requeue).To(gomega.BeFalse())
	g.Expect(result.RequeueAfter).To(gomega.Equal(testCfg.RoleBindingRequeueDuration))

	roleBinding = &rbacv1.RoleBinding{}
	g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseRoleBinding.GetName()}, roleBinding)).To(gomega.Succeed())
	compareRoleBinding(g, roleBinding, roleBindingCopy)

	t.Log("updating the modified RoleBinding in user namespace")
	userCopy := roleBinding.DeepCopy()
	userCopy.Subjects = []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      "testname2",
			Namespace: namespace,
		},
	}
	g.Expect(k8sClient.Update(context.TODO(), userCopy)).To(gomega.Succeed())

	result, err = reconciler.Reconcile(request)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(result.Requeue).To(gomega.BeFalse())
	g.Expect(result.RequeueAfter).To(gomega.Equal(testCfg.RoleBindingRequeueDuration))

	roleBinding = &rbacv1.RoleBinding{}
	g.Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: namespace, Name: baseRoleBinding.GetName()}, roleBinding)).To(gomega.Succeed())
	compareRoleBinding(g, roleBinding, roleBindingCopy)
}

func TestRoleBindingReconciler_predicate(t *testing.T) {
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
		g := gomega.NewGomegaWithT(t)
		deleteEventPod := event.DeleteEvent{Meta: pod.GetObjectMeta(), Object: pod}
		deleteEventLabelledSrvAcc := event.DeleteEvent{Meta: labelledRoleBinding.GetObjectMeta(), Object: labelledRoleBinding}
		deleteEventUnlabelledSrvAcc := event.DeleteEvent{Meta: unlabelledRoleBinding.GetObjectMeta(), Object: unlabelledRoleBinding}

		g.Expect(preds.Delete(deleteEventPod)).To(gomega.BeFalse())
		g.Expect(preds.Delete(deleteEventLabelledSrvAcc)).To(gomega.BeFalse())
		g.Expect(preds.Delete(deleteEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		g.Expect(preds.Delete(event.DeleteEvent{})).To(gomega.BeFalse())
	})

	t.Run("createFunc", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		createEventPod := event.CreateEvent{Meta: pod.GetObjectMeta(), Object: pod}
		createEventLabelledSrvAcc := event.CreateEvent{Meta: labelledRoleBinding.GetObjectMeta(), Object: labelledRoleBinding}
		createEventUnlabelledSrvAcc := event.CreateEvent{Meta: unlabelledRoleBinding.GetObjectMeta(), Object: unlabelledRoleBinding}

		g.Expect(preds.Create(createEventPod)).To(gomega.BeFalse())
		g.Expect(preds.Create(createEventLabelledSrvAcc)).To(gomega.BeTrue())
		g.Expect(preds.Create(createEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		g.Expect(preds.Create(event.CreateEvent{})).To(gomega.BeFalse())
	})

	t.Run("genericFunc", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		genericEventPod := event.GenericEvent{Meta: pod.GetObjectMeta(), Object: pod}
		genericEventLabelledSrvAcc := event.GenericEvent{Meta: labelledRoleBinding.GetObjectMeta(), Object: labelledRoleBinding}
		genericEventUnlabelledSrvAcc := event.GenericEvent{Meta: unlabelledRoleBinding.GetObjectMeta(), Object: unlabelledRoleBinding}

		g.Expect(preds.Generic(genericEventPod)).To(gomega.BeFalse())
		g.Expect(preds.Generic(genericEventLabelledSrvAcc)).To(gomega.BeTrue())
		g.Expect(preds.Generic(genericEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		g.Expect(preds.Generic(event.GenericEvent{})).To(gomega.BeFalse())
	})

	t.Run("updateFunc", func(t *testing.T) {
		g := gomega.NewGomegaWithT(t)
		updateEventPod := event.UpdateEvent{MetaNew: pod.GetObjectMeta(), ObjectNew: pod}
		updateEventLabelledSrvAcc := event.UpdateEvent{MetaNew: labelledRoleBinding.GetObjectMeta(), ObjectNew: labelledRoleBinding}
		updateEventUnlabelledSrvAcc := event.UpdateEvent{MetaNew: unlabelledRoleBinding.GetObjectMeta(), ObjectNew: unlabelledRoleBinding}

		g.Expect(preds.Update(updateEventPod)).To(gomega.BeFalse())
		g.Expect(preds.Update(updateEventUnlabelledSrvAcc)).To(gomega.BeFalse())
		g.Expect(preds.Update(updateEventLabelledSrvAcc)).To(gomega.BeTrue())
		g.Expect(preds.Update(event.UpdateEvent{})).To(gomega.BeFalse())
	})
}
