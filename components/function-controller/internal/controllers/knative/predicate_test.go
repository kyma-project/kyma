package knative

import (
	"testing"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"

	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
)

func TestServiceReconciler_getPredicates(t *testing.T) {
	gm := gomega.NewGomegaWithT(t)

	r := &ServiceReconciler{Log: ctrl.Log}
	preds := r.getPredicates()

	srv := &servingv1.Service{}
	pod := &corev1.Pod{}
	labelledSrv := &servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{
			serverlessv1alpha1.FunctionManagedByLabel: "whatever",
			serverlessv1alpha1.FunctionNameLabel:      "whatever-2",
			serverlessv1alpha1.FunctionUUIDLabel:      "whatever-3",
		}},
	}

	t.Run("deleteFunc should return false on any event", func(t *testing.T) {
		deleteEventPod := event.DeleteEvent{Meta: pod.GetObjectMeta(), Object: pod}
		deleteEventSrv := event.DeleteEvent{Meta: srv.GetObjectMeta(), Object: srv}

		gm.Expect(preds.Delete(deleteEventSrv)).To(gomega.BeFalse())
		gm.Expect(preds.Delete(deleteEventPod)).To(gomega.BeFalse())
		gm.Expect(preds.Delete(event.DeleteEvent{})).To(gomega.BeFalse())
	})

	t.Run("createFunc should return false on any event", func(t *testing.T) {
		createEventPod := event.CreateEvent{Meta: pod.GetObjectMeta(), Object: pod}
		createEventSrv := event.CreateEvent{Meta: srv.GetObjectMeta(), Object: srv}

		gm.Expect(preds.Create(createEventSrv)).To(gomega.BeFalse())
		gm.Expect(preds.Create(createEventPod)).To(gomega.BeFalse())
		gm.Expect(preds.Create(event.CreateEvent{})).To(gomega.BeFalse())
	})

	t.Run("genericFunc should return true on correct event", func(t *testing.T) {
		genericEventPod := event.GenericEvent{Meta: pod.GetObjectMeta(), Object: pod}
		genericEventSrv := event.GenericEvent{Meta: srv.GetObjectMeta(), Object: srv}

		gm.Expect(preds.Generic(genericEventSrv)).To(gomega.BeFalse())
		gm.Expect(preds.Generic(genericEventPod)).To(gomega.BeFalse())
		gm.Expect(preds.Generic(event.GenericEvent{})).To(gomega.BeFalse())

		gm.Expect(preds.Generic(event.GenericEvent{Object: labelledSrv, Meta: labelledSrv.GetObjectMeta()})).To(gomega.BeTrue(), "only case this should be true")
	})

	t.Run("updateFunc should return true on correct event", func(t *testing.T) {
		updateEventPod := event.UpdateEvent{MetaNew: pod.GetObjectMeta(), ObjectNew: pod}
		updateEventSrv := event.UpdateEvent{MetaNew: srv.GetObjectMeta(), ObjectNew: srv}

		gm.Expect(preds.Update(updateEventSrv)).To(gomega.BeFalse())
		gm.Expect(preds.Update(event.UpdateEvent{})).To(gomega.BeFalse())
		gm.Expect(preds.Update(updateEventPod)).To(gomega.BeFalse())

		gm.Expect(preds.Update(event.UpdateEvent{ObjectNew: labelledSrv, MetaNew: labelledSrv.GetObjectMeta()})).To(gomega.BeTrue(), "only case this should be true")
	})
}
