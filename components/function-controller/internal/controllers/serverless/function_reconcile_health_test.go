package serverless

import (
	"context"
	"testing"

	"github.com/onsi/gomega"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestFunctionReconciler_Reconcile_HealthCheck(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	t.Run("should receive health check and notify to health channel", func(t *testing.T) {
		healthCh := make(chan bool)
		reconciler := NewFunctionReconciler(nil, zap.NewNop().Sugar(), FunctionConfig{}, nil, nil, nil, healthCh)

		healthRequest := ctrl.Request{NamespacedName: types.NamespacedName{
			Name: HealthEvent,
		}}

		go func() {
			result, err := reconciler.Reconcile(context.Background(), healthRequest)

			g.Expect(result).To(gomega.Equal(ctrl.Result{}))
			g.Expect(err).To(gomega.BeNil())
		}()

		g.Expect(<-healthCh).To(gomega.BeTrue())
	})

	t.Run("should receive health check and return error after timeout", func(t *testing.T) {
		reconciler := NewFunctionReconciler(nil, zap.NewNop().Sugar(), FunctionConfig{}, nil, nil, nil, make(chan bool))

		healthRequest := ctrl.Request{NamespacedName: types.NamespacedName{
			Name: HealthEvent,
		}}

		result, err := reconciler.Reconcile(context.Background(), healthRequest)

		g.Expect(result).To(gomega.Equal(ctrl.Result{}))
		g.Expect(err).To(gomega.BeNil())
	})
}
