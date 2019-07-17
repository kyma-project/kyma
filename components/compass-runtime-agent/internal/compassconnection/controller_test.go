package compassconnection

import (
	"testing"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/types"
)

const (
	compassConnectionName = "compass-connection"
)

var (
	compassConnectionNamespacedName = types.NamespacedName{
		Name: compassConnectionName,
	}
)

func TestReconcile(t *testing.T) {

	t.Run("should reconcile request", func(t *testing.T) {
		// given
		compassConnection := &v1alpha1.CompassConnection{
			ObjectMeta: v1.ObjectMeta{
				Name: compassConnectionName,
			},
		}

		client := NewFakeClient(compassConnection)

		reconciler := newReconciler(client)

		request := reconcile.Request{
			NamespacedName: compassConnectionNamespacedName,
		}

		// when
		result, err := reconciler.Reconcile(request)

		// then
		require.NoError(t, err)
		require.Empty(t, result)

	})
}
